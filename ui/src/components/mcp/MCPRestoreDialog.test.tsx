import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, expect, it, vi } from 'vitest';
import { mcpApi, type MCPPlan } from '../../api/mcp';
import { I18nProvider } from '../../i18n';
import MCPRestoreDialog from './MCPRestoreDialog';

vi.mock('../../api/mcp', async (load) => ({ ...await load<typeof import('../../api/mcp')>(), mcpApi: { previewRestore: vi.fn(), restore: vi.fn() } }));

const backups = [
  { id: '1790748600000000000-first', target: 'claude', path: '/.claude.json' },
  { id: '1790748500000000000-second', target: 'pi', path: '/work/app/.pi/mcp.json' },
];
const plan = (revision: string, names: string[]): MCPPlan => ({
  revision, sourcePath: '/config.yaml', blocked: false,
  changes: names.map((name) => ({ target: 'pi', path: backups[1].path, name, action: 'restore' })),
});

beforeEach(() => vi.resetAllMocks());

it('keeps the preview viewport and dialog mounted across loading, results, and errors', async () => {
  const user = userEvent.setup();
  let resolveFirst!: (value: MCPPlan) => void;
  let rejectSecond!: (reason: Error) => void;
  vi.mocked(mcpApi.previewRestore)
    .mockReturnValueOnce(new Promise((resolve) => { resolveFirst = resolve; }))
    .mockReturnValueOnce(new Promise((_, reject) => { rejectSecond = reject; }));
  render(<QueryClientProvider client={new QueryClient()}><I18nProvider><MCPRestoreDialog backups={backups} onClose={vi.fn()} onRestored={vi.fn()} /></I18nProvider></QueryClientProvider>);

  const dialog = screen.getByRole('dialog');
  const viewport = screen.getByRole('region', { name: 'Preview' });
  const list = screen.getByRole('radiogroup');
  const [first, second] = screen.getAllByRole('radio');
  list.scrollTop = 80;
  expect(viewport).toHaveClass('h-28', 'overflow-auto');
  expect(viewport).toHaveAttribute('aria-busy', 'true');
  expect(screen.getByRole('button', { name: 'Restore this file' })).toBeDisabled();

  await act(async () => resolveFirst(plan('first', ['docs', 'search', 'files', 'tools', 'context'])));
  expect(await screen.findByText(/restore\s+context/)).toBeInTheDocument();
  expect(screen.getByRole('region', { name: 'Preview' })).toBe(viewport);
  expect(viewport).toHaveAttribute('aria-busy', 'false');

  await user.click(second);
  expect(viewport).toHaveAttribute('aria-busy', 'true');
  expect(screen.getByRole('button', { name: 'Restore this file' })).toBeDisabled();
  expect(screen.queryByText(/restore\s+context/)).not.toBeInTheDocument();
  expect(screen.getByRole('dialog')).toBe(dialog);
  expect(screen.getAllByRole('radio')[0]).toBe(first);
  expect(list.scrollTop).toBe(80);
  expect(second).toHaveFocus();

  await act(async () => rejectSecond(new Error('Preview unavailable')));
  expect(await screen.findByRole('alert')).toHaveTextContent('Preview unavailable');
  expect(screen.getByRole('region', { name: 'Preview' })).toBe(viewport);
  expect(viewport).toHaveAttribute('aria-busy', 'false');
  expect(screen.getByRole('button', { name: 'Restore this file' })).toBeDisabled();
});

it('restores only the selected backup with its completed preview revision', async () => {
  const user = userEvent.setup();
  const restored = vi.fn();
  let resolveSecond!: (value: MCPPlan) => void;
  vi.mocked(mcpApi.previewRestore)
    .mockResolvedValueOnce(plan('first', ['docs']))
    .mockReturnValueOnce(new Promise((resolve) => { resolveSecond = resolve; }));
  vi.mocked(mcpApi.restore).mockResolvedValue({ applied: [], backupIds: [] });
  render(<QueryClientProvider client={new QueryClient()}><I18nProvider><MCPRestoreDialog backups={backups} onClose={vi.fn()} onRestored={restored} /></I18nProvider></QueryClientProvider>);

  const restore = screen.getByRole('button', { name: 'Restore this file' });
  await waitFor(() => expect(restore).toBeEnabled());
  await user.click(screen.getAllByRole('radio')[1]);
  await user.click(restore);
  expect(mcpApi.restore).not.toHaveBeenCalled();
  expect(mcpApi.previewRestore).toHaveBeenLastCalledWith(backups[1].id);

  await act(async () => resolveSecond(plan('second', ['context'])));
  await waitFor(() => expect(restore).toBeEnabled());
  await user.click(restore);
  await waitFor(() => expect(restored).toHaveBeenCalledOnce());
  expect(mcpApi.restore).toHaveBeenCalledExactlyOnceWith(backups[1].id, 'second');
});
