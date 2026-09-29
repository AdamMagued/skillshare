import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { api } from '../api/client';
import type { Target } from '../api/client';
import { mcpApi } from '../api/mcp';
import { ToastProvider } from '../components/Toast';
import { I18nProvider } from '../i18n';
import SyncPage from './SyncPage';

vi.mock('../api/client', async (load) => ({
  ...await load<typeof import('../api/client')>(),
  api: { listTargets: vi.fn(), diff: vi.fn(), diffExtras: vi.fn(), listLog: vi.fn(), skillsOffPreview: vi.fn(), updateTarget: vi.fn() },
}));
vi.mock('../api/mcp', async (load) => ({ ...await load<typeof import('../api/mcp')>(), mcpApi: { list: vi.fn() } }));

const target = (name: string) => ({
  name, path: '/home/me/.agents/skills', mode: 'merge', targetNaming: 'flat', status: 'merged', linkedCount: 3, localCount: 0,
  include: [], exclude: [], expectedSkillCount: 3, skillsEnabled: true,
}) as Target;

describe('Sync page folder conflicts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.listTargets).mockResolvedValue({ targets: [target('codex'), target('universal')], sourceSkillCount: 3 });
    vi.mocked(api.diff).mockResolvedValue({
      diffs: [], ignored_count: 0, ignored_skills: [], ignore_root: '', ignore_repos: [],
      folder_conflicts: [{ path: '/home/me/.agents/skills', targets: ['codex', 'universal'], keep: 'universal', stop: ['codex'] }],
    });
    vi.mocked(api.diffExtras).mockResolvedValue({ extras: [] });
    vi.mocked(api.listLog).mockResolvedValue({ entries: [] } as never);
    vi.mocked(mcpApi.list).mockResolvedValue({ paths: {}, source: { targets: [], servers: {} } } as never);
    vi.mocked(api.skillsOffPreview).mockResolvedValue({ remove: [], keep: [], sharedWith: 'universal' });
    vi.mocked(api.updateTarget).mockResolvedValue({ success: true });
  });

  it('names the targets that undo each other and stops skills for the one to drop', async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <QueryClientProvider client={new QueryClient()}><I18nProvider><ToastProvider><SyncPage /></ToastProvider></I18nProvider></QueryClientProvider>
      </MemoryRouter>,
    );
    expect(await screen.findByText(/codex and universal sync skills to the same folder .*\.agents\/skills with different filters, so each sync undoes the other\./)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Stop syncing skills for codex' }));
    await user.click(await screen.findByRole('button', { name: 'Stop syncing' }));
    await waitFor(() => expect(api.updateTarget).toHaveBeenCalledWith('codex', { skills_enabled: false }));
  });
});
