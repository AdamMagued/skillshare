import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { I18nProvider } from '../../i18n';
import MCPServerList from './MCPServerList';

describe('MCP server list', () => {
  it('shows which Pi adapter settings a server has on its row, without their contents', () => {
    const server = { command: 'npx', targets: ['pi'], piExtension: 'pi-mcp-adapter', directTools: ['take_screenshot', 'list_pages'], piOptions: { excludeTools: ['secret_*'] } };
    render(<I18nProvider><MCPServerList rows={[{ name: 'docs', server, cells: {} }]} targets={['pi']} targetsOf={() => ['pi']} onToggle={vi.fn()} onMenu={vi.fn()} /></I18nProvider>);
    const line = screen.getByText('take_screenshot, list_pages').parentElement!;
    expect(line).toHaveTextContent('Direct tools');
    expect(line).toHaveTextContent('Other Pi settings');
    expect(line).not.toHaveTextContent('secret_');
  });
});
