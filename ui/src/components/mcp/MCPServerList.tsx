import { Fragment, useState } from 'react';
import { Ellipsis, Plug } from 'lucide-react';
import { useT } from '../../i18n';
import AgentIcon from '../AgentIcon';
import { mcpOffTargets } from '../../api/mcp';
import { describeEndpoint, writes, type MatrixRow } from './mcpView';
import { TargetPill, TargetToggles } from './TargetPicker';
import { useDirectToolsLabel } from './MCPProjectSettings';

interface Props {
  rows: MatrixRow[];
  targets: string[];
  targetsOf: (name: string) => string[];
  onToggle: (name: string, target: string, on: boolean) => void;
  onMenu: (e: React.MouseEvent<HTMLButtonElement>, name: string) => void;
  /** Agents a switch-only entry can go to. mcp.projects cannot reach Claude's off list. */
  offTargets?: readonly string[];
  /** Holds the toggles while a save is on its way: each sends the revision it previewed. */
  disabled?: boolean;
}

/**
 * One row per server. The agents it writes to are chips inside the row, not columns:
 * the list grows downwards as more CLIs gain MCP support, so it never scrolls sideways.
 */
export default function MCPServerList({ rows, targets, targetsOf, onToggle, onMenu, offTargets = mcpOffTargets, disabled = false }: Props) {
  const t = useT();
  const [open, setOpen] = useState<string[]>([]);
  const directToolsLabel = useDirectToolsLabel();

  return (
    <div className="ss-list">
      {rows.map((row) => {
        // A switch-only entry works in a few Agents, so it offers and counts only those.
        const offered = row.server?.disabled ? targets.filter((x) => offTargets.includes(x)) : targets;
        const selected = row.server ? targetsOf(row.name).filter((x) => offered.includes(x)) : [];
        const expanded = open.includes(row.name);
        const http = Boolean(row.server?.url);
        const direct = row.server?.directTools;
        const piOptions = Object.keys(row.server?.piOptions ?? {}).length > 0;
        const piSettings = Boolean(row.server?.piExtension) && selected.includes('pi');
        return (
          <Fragment key={row.name}>
            <div className="ss-r">
              <span className="ss-cat sm mcp"><Plug size={14} /></span>
              <span className="flex min-w-0 flex-1 flex-col gap-0.5">
                <span className="flex items-center gap-2">
                  <span title={row.name} className={`truncate font-mono font-semibold ${row.server ? '' : 'text-ink-3 line-through'}`}>{row.name}</span>
                  {row.server && Object.values(row.cells).some(writes) && <span className="ss-tag warn">{t('plugins.pending')}</span>}
                  {row.server && !row.server.disabled && row.server.targets?.length === 0 && <span className="ss-tag">{t('plugins.noAgentsYet')}</span>}
                  {!row.server && <span className="ss-st bad">{t('mcp.removedFromSource')}</span>}
                </span>
                {/* Transport and endpoint on one quiet line, the same shape as a plugin row. */}
                <span className="truncate text-xs text-ink-3">{!row.server ? t('mcp.removedHint') : row.server.disabled ? t('mcp.offHere') : <>{http ? 'http' : 'stdio'} · <span className="font-mono">{describeEndpoint(row.server)}</span></>}</span>
                {/* Only whether other Pi settings exist, never their contents: piOptions may hold anything. */}
                {piSettings && (
                  <span className="flex min-w-0 items-center gap-1.5 whitespace-nowrap text-xs text-ink-2">
                    <AgentIcon target="pi" size={12} />
                    <span>{row.server?.piExtension === 'builtin' ? t('mcp.piBuiltinShort') : row.server?.piExtension}</span>
                    {row.server?.piExtension === 'builtin' && <span>· {String(row.server.piOptions?.exposure ?? 'codemode')}</span>}
                    {direct !== undefined && <><span className="text-ink-3">{t('mcp.directTools')}</span><span className={`truncate ${Array.isArray(direct) ? 'font-mono' : ''}`} title={directToolsLabel(direct)}>{directToolsLabel(direct)}</span></>}
                    {direct !== undefined && piOptions && <span className="text-ink-3">·</span>}
                    {piOptions && <span>{t('mcp.piOptions')}</span>}
                  </span>
                )}
              </span>
              {row.server && (
                <>
                  <TargetPill selected={selected} text={`${selected.length}/${offered.length}`} expanded={expanded} label={t('mcp.chooseAgents', { name: row.name })} onClick={() => setOpen((prev) => (expanded ? prev.filter((x) => x !== row.name) : [...prev, row.name]))} />
                  <button type="button" className="ss-ib" aria-label={t('mcp.moreActions', { name: row.name })} onClick={(e) => onMenu(e, row.name)}>
                    <Ellipsis size={16} />
                  </button>
                </>
              )}
            </div>
            {expanded && row.server && (
              <div className="ss-r fold !min-h-0 flex-wrap gap-x-6 gap-y-3.5 !py-3.5">
                <TargetToggles offered={offered} selected={selected} http={http} disabled={disabled} onToggle={(target, on) => onToggle(row.name, target, on)} />
              </div>
            )}
          </Fragment>
        );
      })}
    </div>
  );
}
