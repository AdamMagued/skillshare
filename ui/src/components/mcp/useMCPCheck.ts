import { useState } from 'react';
import { mcpCheckApi } from '../../api/mcpCheck';
import type { MCPCheckFinding, MCPCheckReport, MCPCheckServer } from '../../api/mcpCheck';

/** The servers the MCP page lists. A project server may share a global name, so it never counts here. */
export function globalServers(report?: MCPCheckReport): MCPCheckServer[] {
  return (report?.servers ?? []).filter((server) => !server.project);
}

/** What the page shows of a check: errors and warnings. Info findings stay in the CLI. */
export function problemsByServer(report?: MCPCheckReport): Record<string, MCPCheckFinding[]> {
  const out: Record<string, MCPCheckFinding[]> = {};
  for (const server of globalServers(report)) {
    const problems = server.findings.filter((f) => f.level !== 'info');
    if (problems.length > 0) out[server.name] = problems;
  }
  return out;
}

/** Runs the check on demand only. The result lives in memory, so a reload clears it. */
export function useMCPCheck() {
  const [result, setResult] = useState<{ report?: MCPCheckReport; checkedAt?: number; error?: string }>({});
  const [running, setRunning] = useState(false);
  const run = async () => {
    setRunning(true);
    try {
      setResult({ report: await mcpCheckApi.run(), checkedAt: Date.now() });
    } catch (e) {
      setResult({ error: (e as Error).message });
    } finally {
      setRunning(false);
    }
  };
  return { ...result, running, run };
}
