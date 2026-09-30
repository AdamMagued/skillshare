import { apiFetch } from './client';

/** One result of `skillshare mcp check`. `subject` names the variable, command, or host; never a value. */
export interface MCPCheckFinding {
  level: 'error' | 'warning' | 'info';
  check: string;
  target: string;
  message: string;
  subject?: string;
}
export interface MCPCheckServer { name: string; ok: boolean; findings: MCPCheckFinding[] }
export interface MCPCheckReport { servers: MCPCheckServer[]; summary: { errors: number; warnings: number } }

export const mcpCheckApi = {
  /** Read-only. Findings that are errors still resolve; only a check that cannot run rejects. */
  run: () => apiFetch<MCPCheckReport>('/mcp/check'),
};
