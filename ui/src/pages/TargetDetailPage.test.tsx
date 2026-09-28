import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { api, type Target } from '../api/client';
import { mcpApi } from '../api/mcp';
import { ToastProvider } from '../components/Toast';
import { I18nProvider } from '../i18n';
import TargetDetailPage from './TargetDetailPage';

vi.mock('../api/client', async (load) => ({
  ...await load<typeof import('../api/client')>(),
  api: {
    listTargets: vi.fn(), updateTarget: vi.fn(), availableTargets: vi.fn(), getTargetInstructions: vi.fn(),
    previewSyncMatrix: vi.fn(), listExtraExtensions: vi.fn(),
  },
}));
// The file editor needs a data router; these tests look at the page around it.
vi.mock('../components/instructions/TargetInstructions', () => ({ default: () => null }));
vi.mock('../api/mcp', async (load) => ({ ...await load<typeof import('../api/mcp')>(), mcpApi: { list: vi.fn() } }));

const target = (over: Partial<Target>) => ({
  path: '/home/me/.gemini/skills', mode: 'merge', targetNaming: 'flat', status: 'merged', linkedCount: 0, localCount: 0,
  include: [], exclude: [], expectedSkillCount: 0, skillsEnabled: true, ...over,
}) as Target;
const view = (name: string, targets: Target[]) => {
  vi.mocked(api.listTargets).mockResolvedValue({ targets, sourceSkillCount: 9 });
  render(
    <MemoryRouter initialEntries={[`/targets/${name}`]}>
      <QueryClientProvider client={new QueryClient()}><I18nProvider><ToastProvider>
        <Routes><Route path="/targets/:name" element={<TargetDetailPage />} /></Routes>
      </ToastProvider></I18nProvider></QueryClientProvider>
    </MemoryRouter>,
  );
};

describe('Target detail skills switch', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.availableTargets).mockResolvedValue({ targets: [] });
    vi.mocked(api.getTargetInstructions).mockResolvedValue({ supported: true, path: '/home/me/.gemini/GEMINI.md', read_order: [] } as never);
    vi.mocked(api.previewSyncMatrix).mockResolvedValue({ entries: [] });
    vi.mocked(api.listExtraExtensions).mockResolvedValue({ extensions: [] });
    vi.mocked(mcpApi.list).mockResolvedValue({ paths: {}, source: { targets: [], servers: {} } } as never);
    vi.mocked(api.updateTarget).mockResolvedValue({ success: true });
  });

  it('shows a target with skills off as not synced, naming the folder it reads, and resumes it', async () => {
    const user = userEvent.setup();
    view('gemini', [
      target({ name: 'gemini', skillsEnabled: false, skillsReadFrom: ['universal'] }),
      target({ name: 'universal', path: '/home/me/.agents/skills', linkedCount: 9, skillsAlsoReadBy: ['gemini'] }),
    ]);
    expect(await screen.findByText('Skills don’t sync to gemini')).toBeInTheDocument();
    expect(screen.getByText(/gemini still reads 9 skills from universal’s ~\/.agents\/skills/)).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Save' })).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Resume syncing' }));
    await waitFor(() => expect(api.updateTarget).toHaveBeenCalledWith('gemini', { skills_enabled: true }));
  });

  it('links the targets that read this skills folder instead of syncing their own', async () => {
    view('universal', [
      target({ name: 'gemini', skillsEnabled: false, skillsReadFrom: ['universal'] }),
      target({ name: 'universal', path: '/home/me/.agents/skills', linkedCount: 9, skillsAlsoReadBy: ['gemini'] }),
    ]);
    expect(await screen.findByRole('link', { name: 'gemini' })).toHaveAttribute('href', '/targets/gemini');
    expect(screen.getByRole('button', { name: 'Stop syncing skills' })).toBeInTheDocument();
  });

  it('names the targets a tool with skills on also reads, since it sees their skills twice', async () => {
    vi.mocked(api.availableTargets).mockResolvedValue({ targets: [{ name: 'pi', path: '/home/me/.pi/agent/skills', installed: true, detected: false, readsFrom: ['universal'] }] });
    view('pi', [
      target({ name: 'pi', path: '/home/me/.pi/agent/skills' }),
      target({ name: 'universal', path: '/home/me/.agents/skills', linkedCount: 9 }),
    ]);
    expect(await screen.findByText(/This target also reads the skills of/)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'universal' })).toHaveAttribute('href', '/targets/universal');
  });

  it('shows the instruction file path under the title, like the other tabs', async () => {
    vi.mocked(api.getTargetInstructions).mockResolvedValue({ supported: true, target: 'gemini', path: '/home/me/.gemini/GEMINI.md', exists: true, content: '', read_order: [], convert: [], shared: [] } as never);
    view('gemini?tab=instructions', [target({ name: 'gemini' })]);
    expect(await screen.findByRole('heading', { name: 'gemini' })).toBeInTheDocument();
    expect(await screen.findByText('~/.gemini/GEMINI.md')).toBeInTheDocument();
  });
});
