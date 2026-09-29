import { api, type DiffTarget, type ExtraDiffResult, type SyncResponse, type Target } from '../../api/client';
import { mcpApi, type MCPPlan } from '../../api/mcp';
import { formatAgentDisplayName } from '../../lib/resourceNames';
import { groupByFile, projectOf, type MCPChange } from '../mcp/mcpView';

export type Part = 'skill' | 'agent' | 'extra' | 'mcp';
export type RowIcon = 'add' | 'adopt' | 'update' | 'remove' | 'kept' | 'conflict';

export interface ChangeRow {
  key: string;
  part: Part;
  name: string;
  icon: RowIcon;
  /** i18n key, or null when `detail` is backend text shown as is */
  text: string | null;
  detail?: string;
  /** False for rows sync leaves alone (kept local copies, conflicts) */
  counts: boolean;
  /** Edited inside the target: sync keeps it unless Force is on */
  edited?: boolean;
  /** An MCP switch: added, it turns a global server off in a project; removed, back on */
  switch?: boolean;
}

export interface ChangeGroup {
  key: string;
  part: 'target' | 'extra' | 'mcp';
  name: string;
  mode?: string;
  path?: string;
  /** Set on Claude Code's off list, which sits in the global file: the project it is for */
  project?: string;
  rows: ChangeRow[];
}

const editedRow = (force: boolean) => ({ icon: force ? 'update' : 'kept', text: force ? 'sync.row.forceReplace' : 'sync.row.kept', counts: force, edited: true }) as const;

const flatKey = (name: string) => name.replace(/\//g, '__').replace(/\.md$/i, '');

/**
 * One group per target with pending skill or agent changes, in name order (the API order changes between requests).
 * Local-only items are listed elsewhere. `ignored` names the .skillignore/.agentignore matches, which prune like deleted items.
 */
export function resourceGroups(diffs: DiffTarget[], targets: Target[], parts: Set<Part>, force: boolean, ignored: { skill?: string[]; agent?: string[] } = {}): { groups: ChangeGroup[]; inSync: string[] } {
  const groups: ChangeGroup[] = [];
  const inSync: string[] = [];
  const ignoredKeys = { skill: new Set((ignored.skill ?? []).map(flatKey)), agent: new Set((ignored.agent ?? []).map(flatKey)) };
  for (const d of [...diffs].sort((a, b) => a.target.localeCompare(b.target))) {
    const target = targets.find((t) => t.name === d.target);
    const rows: ChangeRow[] = [];
    for (const item of d.items ?? []) {
      const part = item.kind === 'agent' ? 'agent' : 'skill';
      if (!parts.has(part) || item.action === 'local') continue;
      // Agents always sync as symlinks, whatever the target mode.
      const copy = part === 'skill' && target?.mode === 'copy';
      const row = { key: `${d.target}/${part}/${item.skill}`, part, name: part === 'agent' ? formatAgentDisplayName(item.skill) : item.skill } as const;
      if (item.skill === '(target naming)') rows.push({ ...row, icon: 'kept', text: null, detail: item.reason, counts: false });
      else if (item.skill === '(entire directory)') rows.push({ ...row, name: target?.path ?? item.skill, icon: 'add', text: 'sync.row.folder', counts: true });
      else if (item.action === 'link') rows.push({ ...row, icon: 'add', text: item.reason?.startsWith('missing') ? 'sync.row.recopy' : copy ? 'sync.row.newCopy' : 'sync.row.newLink', counts: true });
      else if (item.action === 'update') rows.push({ ...row, icon: 'update', text: copy ? 'sync.row.updateCopy' : 'sync.row.updateLink', counts: true });
      else if (item.action === 'prune') rows.push({ ...row, icon: 'remove', text: ignoredKeys[part].has(flatKey(item.skill)) ? 'sync.row.pruneIgnored' : 'sync.row.prune', counts: true });
      else if (item.action === 'skip') rows.push({ ...row, ...editedRow(force) });
    }
    if (rows.length) groups.push({ key: `target/${d.target}`, part: 'target', name: d.target, mode: target?.mode, path: target?.path, rows });
    else inSync.push(d.target);
  }
  return { groups, inSync };
}

// A regular file where a link belongs, or a copy that differs, is only overwritten with force (syncOneExtraFile).
const EXTRA_EDITED = new Set(['not a symlink', 'content differs']);

export function extraGroups(diffs: ExtraDiffResult[], force: boolean): ChangeGroup[] {
  return diffs.filter((d) => d.items?.length).map((d) => ({
    key: `extra/${d.name}/${d.target}`,
    part: 'extra',
    name: d.name,
    mode: d.mode,
    path: d.target,
    // Without a source folder the sync only creates it, so nothing lands in the target.
    rows: d.items.map((i) => (i.reason === 'no source directory'
      ? { key: `${d.name}/${d.target}/*`, part: 'extra', name: d.name, icon: 'kept', text: 'sync.row.extraNoSource', counts: false }
      : {
        key: `${d.name}/${d.target}/${i.file}`,
        part: 'extra',
        name: i.file,
        ...(EXTRA_EDITED.has(i.reason) ? editedRow(force) : {
          icon: i.action === 'update' ? 'update' : 'add',
          text: i.action === 'update' ? 'sync.row.extraUpdate' : 'sync.row.extraCreate',
          counts: true,
        }),
      })),
  }));
}

const MCP_ICON: Record<string, RowIcon> = { add: 'add', adopt: 'adopt', update: 'update', remove: 'remove', conflict: 'conflict' };

export function mcpGroups(plan: MCPPlan | null | undefined): ChangeGroup[] {
  return groupByFile((plan?.changes ?? []).filter((c) => MCP_ICON[c.action])).map((file) => ({
    key: `mcp/${file.key}`,
    part: 'mcp',
    name: file.target,
    path: file.path,
    project: file.offListFor,
    rows: file.changes.map((c: MCPChange) => ({
      key: `${file.key}/${c.name}`,
      part: 'mcp',
      name: c.name,
      icon: MCP_ICON[c.action],
      // A switch-only entry adds or removes no server: it turns one off for a project, or back on.
      text: c.action === 'conflict' ? null : c.switch && c.action !== 'update' ? `sync.row.mcp.switch.${c.action}` : `sync.row.mcp.${c.action}`,
      detail: c.message,
      counts: c.action !== 'conflict',
      switch: c.switch,
    })),
  }));
}

/** Targets already in sync, the global ones apart from each project's: a project target is `<project>@<tool>`. */
export function groupInSync(names: string[]) {
  const global: string[] = [];
  const projects = new Map<string, string[]>();
  for (const name of names) {
    const at = name.lastIndexOf('@');
    if (at < 0) global.push(name);
    else projects.set(name.slice(0, at), [...(projects.get(name.slice(0, at)) ?? []), name.slice(at + 1)]);
  }
  return { global, projects: [...projects].map(([project, tools]) => ({ project, tools })) };
}

/** Ignored names by folder, so a shared prefix such as `security/` is written once. */
export function groupByFolder(names: string[]) {
  const folders = new Map<string, string[]>();
  for (const name of names) {
    const slash = name.lastIndexOf('/');
    const folder = slash < 0 ? '' : name.slice(0, slash + 1);
    folders.set(folder, [...(folders.get(folder) ?? []), name.slice(slash + 1)]);
  }
  return [...folders].map(([folder, items]) => ({ folder, items }));
}

export const countChanges = (groups: ChangeGroup[]) => groups.reduce((n, g) => n + g.rows.filter((r) => r.counts).length, 0);
/** Changes a plain sync of every part would apply (the sidebar badge). A blocked MCP plan applies nothing. */
export function pendingCount(diffs: DiffTarget[], targets: Target[], extras: ExtraDiffResult[], plan: MCPPlan | null | undefined): number {
  const parts = new Set<Part>(['skill', 'agent', 'extra', 'mcp']);
  return countChanges([...resourceGroups(diffs, targets, parts, false).groups, ...extraGroups(extras, false), ...(plan?.blocked ? [] : mcpGroups(plan))]);
}
export const countEdited = (groups: ChangeGroup[]) => groups.reduce((n, g) => n + g.rows.filter((r) => r.edited).length, 0);

/** Thrown when the MCP plan no longer matches the one on screen. */
export const MCP_CHANGED = 'mcp-changed';

/** The plan's changes for one mcp.projects root. */
export const projectChanges = (plan: MCPPlan | null | undefined, root: string) => (plan?.changes ?? []).filter((c) => projectOf([root], c) === root);

// Revisions move on any config write, including the resource sync; only the reviewed changes must hold.
const changeKey = (plan: MCPPlan, root?: string) => (root ? projectChanges(plan, root) : plan.changes).filter((c) => c.action !== 'unchanged').map((c) => JSON.stringify([c.target, c.name, c.action])).sort().join();

export interface SyncRun {
  resources: 'skill' | 'agent' | 'both' | null;
  extras: boolean;
  /** The MCP plan the user reviewed, or null to leave MCP alone */
  mcp: MCPPlan | null;
  force: boolean;
  /** Write one project only: `root` as declared under projects, `path` its folder (the mcp.projects key) */
  project?: { root: string; path: string };
}

/** Writes each included part in order. MCP is checked before anything is written and again right before it applies. */
export async function runSync(run: SyncRun) {
  const reviewed = run.mcp;
  const recheck = async () => {
    const fresh = await mcpApi.preview();
    const root = run.project?.path;
    const blocked = root ? projectChanges(fresh, root).some((c) => c.action === 'conflict') : fresh.blocked;
    if (blocked || changeKey(fresh, root) !== changeKey(reviewed!, root)) throw new Error(MCP_CHANGED);
    return fresh.revision;
  };
  if (reviewed) await recheck();
  let resources: SyncResponse | undefined;
  if (run.resources) {
    resources = await api.sync({ force: run.force, ...(run.resources !== 'both' && { kind: run.resources }), ...(run.project && { project: run.project.root }) });
  }
  if (run.extras) {
    const extras = await api.syncExtras({ force: run.force });
    const failed = extras.extras.flatMap((e) => e.targets).find((e) => e.error || e.errors?.length);
    if (failed) throw new Error(failed.error || failed.errors?.join('; '));
  }
  if (reviewed) {
    if (run.project) await mcpApi.syncProject(run.project.path, await recheck());
    else await mcpApi.configure({}, await recheck(), true);
  }
  return { resources };
}
