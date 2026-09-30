import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { MCPCheckFinding, MCPCheckReport } from '../../api/mcpCheck';
import { I18nProvider } from '../../i18n';
import MCPCheckFindings, { MCPCheckTag } from './MCPCheckFindings';
import MCPCheckNote from './MCPCheckNote';

const env: MCPCheckFinding = { level: 'error', check: 'env', target: '', subject: 'CLICKUP_TOKEN', message: 'bearerToken reads CLICKUP_TOKEN, which is not set' };
const rule: MCPCheckFinding = { level: 'error', check: 'client-rule', target: 'claude-desktop', message: 'claude-desktop accepts stdio servers only' };
const shadow: MCPCheckFinding = { level: 'warning', check: 'sync', target: 'claude', message: 'a local scope server wins' };
const report = (servers: MCPCheckReport['servers'], errors = 0, warnings = 0): MCPCheckReport => ({ servers, summary: { errors, warnings } });

const note = (r: MCPCheckReport) => render(<I18nProvider><MCPCheckNote report={r} checkedAt={Date.now()} running={false} onRun={vi.fn()} /></I18nProvider>);
const rows = (findings: MCPCheckFinding[]) => render(<I18nProvider><MCPCheckFindings findings={findings} /></I18nProvider>);

describe('MCP check note', () => {
  it('counts the servers with problems and the errors and warnings', () => {
    note(report([{ name: 'clickup', ok: false, findings: [env, rule] }, { name: 'figma', ok: true, findings: [shadow] }, { name: 'parked', ok: true, findings: [{ level: 'info', check: 'targets', target: '', message: 'kept' }] }], 2, 1));
    expect(screen.getByText('2 servers have problems').closest('.ss-note')).toHaveTextContent('2 error, 1 warning · checked just now');
  });

  it('says every server is fine when there is no error or warning', () => {
    note(report([{ name: 'a', ok: true, findings: [] }, { name: 'b', ok: true, findings: [{ level: 'info', check: 'sync', target: 'claude', message: 'in sync' }] }]));
    expect(screen.getByText('2 servers have no problems').closest('.ss-note')).toHaveClass('inf');
  });
});

describe('MCP check findings', () => {
  it('shows the Agent a finding is about, and no Agent for a server-wide one', () => {
    const { container } = rows([env, rule]);
    const lines = container.querySelectorAll('.ss-r.fold > div');
    expect([lines[0].querySelector('.ss-at'), lines[1].querySelector('.ss-at')].map(Boolean)).toEqual([false, true]);
  });

  it('phrases a missing variable from its subject', () => {
    rows([env]);
    expect(screen.getByText('Reads CLICKUP_TOKEN, but this variable is not set.')).toBeInTheDocument();
  });

  it('keeps the server message for checks without a sentence of their own', () => {
    rows([rule]);
    expect(screen.getByText('claude-desktop accepts stdio servers only')).toBeInTheDocument();
  });

  it('tags a server with warnings only by its warning count', () => {
    render(<I18nProvider><MCPCheckTag findings={[shadow]} /></I18nProvider>);
    expect(screen.getByText('1 warning')).toHaveClass('warn');
  });
});
