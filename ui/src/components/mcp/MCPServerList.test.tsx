import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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

  it('opens and closes the target toggles from the target count', async () => {
    const user = userEvent.setup();
    render(<I18nProvider><MCPServerList rows={[{ name: 'docs', server: { command: 'npx' }, cells: {} }]} targets={['claude', 'cursor']} targetsOf={() => ['claude']} onToggle={vi.fn()} onMenu={vi.fn()} /></I18nProvider>);
    const edit = screen.getByRole('button', { name: 'Choose which agents get docs' });
    await user.click(edit);
    expect(screen.getByRole('checkbox', { name: /Cursor/ })).toBeInTheDocument();
    await user.click(edit);
    expect(screen.queryByRole('checkbox', { name: /Cursor/ })).not.toBeInTheDocument();
  });
});
