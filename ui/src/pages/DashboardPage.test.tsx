import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { api } from '../api/client';
import type { Overview, Target } from '../api/client';
import { mcpApi } from '../api/mcp';
import { pluginsApi } from '../api/plugins';
import { ToastProvider } from '../components/Toast';
import { I18nProvider } from '../i18n';
import DashboardPage from './DashboardPage';

vi.mock('../context/AppContext', () => ({ useAppContext: () => ({ isProjectMode: false }) }));
vi.mock('../api/client', async (load) => {
  const actual = await load<typeof import('../api/client')>();
  return {
    ...actual,
    api: {
      ...actual.api,
      getOverview: vi.fn(), listTargets: vi.fn(), listExtras: vi.fn(), listLog: vi.fn(),
      auditAll: vi.fn(), check: vi.fn(), getVersionCheck: vi.fn(),
    },
  };
});
vi.mock('../api/mcp', async (load) => {
  const actual = await load<typeof import('../api/mcp')>();
  return { ...actual, mcpApi: { ...actual.mcpApi, list: vi.fn() } };
});
vi.mock('../api/plugins', async (load) => {
  const actual = await load<typeof import('../api/plugins')>();
  return { ...actual, pluginsApi: { ...actual.pluginsApi, list: vi.fn() } };
});

function target(name: string, status: string): Target {
  return { name, path: `/home/dev/.${name}/skills`, mode: 'merge', targetNaming: 'flat', status, skillsEnabled: true, linkedCount: 3, localCount: 0, include: [], exclude: [], expectedSkillCount: 3 } as Target;
}

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <MemoryRouter initialEntries={['/']}>
      <QueryClientProvider client={client}><I18nProvider><ToastProvider><DashboardPage /></ToastProvider></I18nProvider></QueryClientProvider>
    </MemoryRouter>,
  );
}

describe('DashboardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.getOverview).mockResolvedValue({ skillCount: 3, agentCount: 0, source: '/home/dev/skills', trackedRepos: [] } as unknown as Overview);
    vi.mocked(api.listTargets).mockResolvedValue({ targets: [target('codex', 'merged'), target('cursor', 'not exist')], sourceSkillCount: 3 } as Awaited<ReturnType<typeof api.listTargets>>);
    vi.mocked(api.listExtras).mockResolvedValue({ extras: [] } as unknown as Awaited<ReturnType<typeof api.listExtras>>);
    vi.mocked(api.listLog).mockResolvedValue({ entries: [] } as unknown as Awaited<ReturnType<typeof api.listLog>>);
    vi.mocked(api.auditAll).mockReturnValue(new Promise(() => {}));
    vi.mocked(api.check).mockReturnValue(new Promise(() => {}));
    vi.mocked(api.getVersionCheck).mockReturnValue(new Promise(() => {}));
    vi.mocked(mcpApi.list).mockReturnValue(new Promise(() => {}));
    vi.mocked(pluginsApi.list).mockReturnValue(new Promise(() => {}));
  });

  it('opens a target row on that target, not the target list', async () => {
    renderPage();
    const links = await screen.findAllByRole('link', { name: /codex/ });
    expect(links.map((a) => a.getAttribute('href'))).toEqual(links.map(() => '/targets/codex'));
  });

  it('opens a broken target from the attention list on that target', async () => {
    renderPage();
    expect(await screen.findByRole('link', { name: 'Open' })).toHaveAttribute('href', '/targets/cursor');
  });
});
