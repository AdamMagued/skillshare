// Extras types
export interface ExtraTarget {
  path: string;
  mode: string;
  flatten: boolean;
  extension?: string;  // transform extension name; presence implies copy mode
  as?: string; // single-file extra: target filename
  status: string;  // "synced" | "drift" | "modified" | "not synced" | "no source"
}

export interface Extra {
  name: string;
  file?: string; // single-file extra: the file in source_dir (AGENTS.md ones are shared instruction files)
  source_dir: string;
  source_type: "per-extra" | "extras_source" | "default";
  file_count: number;
  source_exists: boolean;
  targets: ExtraTarget[];
}

export interface ExtensionInfo {
  name: string;
  description?: string;
  builtin: boolean;
  installed: boolean;
  used_by?: string[]; // names of extras referencing this extension
}

export interface ExtraDiffItem {
  action: string;  // "create" | "update" | "prune"
  file: string;
  reason: string;
}

export interface ExtraDiffResult {
  name: string;
  target: string;
  mode: string;
  synced: boolean;
  items: ExtraDiffItem[];
}

export interface ExtrasSyncResult {
  name: string;
  targets: Array<{
    target: string;
    mode: string;
    synced: number;
    skipped: number;
    pruned: number;
    errors?: string[];
    error?: string;
  }>;
}
