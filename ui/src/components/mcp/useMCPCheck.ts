import { useState } from 'react';
import { mcpCheckApi, type MCPCheckFinding, type MCPCheckReport } from '../../api/mcpCheck';

/** What the page shows of a check: errors and warnings. Info findings stay in the CLI. */
export function problemsByServer(report?: MCPCheckReport): Record<string, MCPCheckFinding[]> {
  const out: Record<string, MCPCheckFinding[]> = {};
  for (const server of report?.servers ?? []) {
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
