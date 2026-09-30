import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { mcpTargets, type MCPServer } from '../../api/mcp';
import { I18nProvider } from '../../i18n';
import MCPDefaults from './MCPDefaults';
import { MCPTargetOrder } from './targetOrder';

const renderDefaults = (targets: string[], onSave = vi.fn(), servers: Record<string, MCPServer> = {}, accounts?: Record<string, { agent: string; configDir: string }>) =>
  render(<I18nProvider><MCPDefaults targets={targets} servers={servers} accounts={accounts} directTools="search" offered={['claude', 'opencode', 'pi']} onSave={onSave} /></I18nProvider>);

describe('MCP defaults', () => {
  it.each([
    ['builtin', { piExtension: 'builtin' }, false],
    ['extension', { piExtension: 'pi-mcp-extension' }, false],
    ['adapter inheriting Pi', { piExtension: 'pi-mcp-adapter' }, true],
    ['adapter targeting another Agent', { piExtension: 'pi-mcp-adapter', targets: ['claude'] }, false],
    ['disabled adapter', { piExtension: 'pi-mcp-adapter', disabled: true }, false],
  ] satisfies [string, MCPServer, boolean][])('offers Direct tools only for an active adapter: %s', (_, server, visible) => {
    renderDefaults(['claude', 'pi'], vi.fn(), { docs: { command: 'npx', ...server } });
    expect(Boolean(screen.queryByText('Direct tools · pi-mcp-adapter'))).toBe(visible);
  });

  it('hides adapter defaults when no server uses them', () => {
    renderDefaults(['pi']);
    expect(screen.queryByText('Direct tools · pi-mcp-adapter')).not.toBeInTheDocument();
  });

  it('offers adapter defaults for explicit Pi targets even without Pi in default targets', () => {
    renderDefaults(['claude'], vi.fn(), { docs: { command: 'npx', piExtension: 'pi-mcp-adapter', targets: ['pi'] } });
    expect(screen.getByText('Direct tools · pi-mcp-adapter')).toBeInTheDocument();
  });

  it('recognizes accounts of Pi as adapter targets', () => {
    renderDefaults(['pi-work'], vi.fn(), { docs: { command: 'npx', piExtension: 'pi-mcp-adapter' } }, { 'pi-work': { agent: 'pi', configDir: '/pi-work' } });
    expect(screen.getByText('Direct tools · pi-mcp-adapter')).toBeInTheDocument();
  });

  it('saves both settings when a target is ticked, since a save replaces both', async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderDefaults(['claude'], onSave);
    expect(screen.queryByText('Direct tools · pi-mcp-adapter')).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Choose the default Agents' }));
    await user.click(screen.getByRole('checkbox', { name: 'Pi' }));
    expect(onSave).toHaveBeenCalledWith({ targets: ['claude', 'pi'], directTools: 'search' });
  });

  // claude-work is a target that is another account of Claude; the page adds such names to the order.
  it('offers an account of an Agent next to the Agents', async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    render(<I18nProvider><MCPTargetOrder.Provider value={[...mcpTargets, 'claude-work']}><MCPDefaults targets={['claude']} servers={{}} directTools={undefined} offered={['claude', 'claude-work']} onSave={onSave} /></MCPTargetOrder.Provider></I18nProvider>);
    await user.click(screen.getByRole('button', { name: 'Choose the default Agents' }));
    await user.click(screen.getByRole('checkbox', { name: 'claude-work' }));
    expect(onSave).toHaveBeenCalledWith({ targets: ['claude', 'claude-work'], directTools: undefined });
  });
});
