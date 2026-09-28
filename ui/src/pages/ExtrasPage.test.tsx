import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { api } from '../api/client';
import { I18nProvider } from '../i18n';
import { ToastProvider } from '../components/Toast';
import ExtrasPage from './ExtrasPage';

vi.mock('../context/AppContext', () => ({ useAppContext: () => ({ isProjectMode: true }) }));
vi.mock('../api/client', async (load) => {
  const actual = await load<typeof import('../api/client')>();
  return {
    ...actual,
    api: {
      ...actual.api,
      listExtras: vi.fn().mockResolvedValue({
        extras: [{ name: 'team-rules', file: 'AGENTS.md', source_dir: '/p/.skillshare/extras/team-rules', source_type: 'per-extra', file_count: 1, source_exists: true,
          targets: [{ path: '/p', mode: 'symlink', flatten: false, as: 'CLAUDE.md', status: 'synced' }] }],
      }),
      listExtraExtensions: vi.fn().mockResolvedValue({ extensions: [] }),
      availableTargets: vi.fn().mockResolvedValue({ targets: [] }),
      getOverview: vi.fn().mockResolvedValue({}),
      addExtraTarget: vi.fn().mockResolvedValue({ success: true }),
    },
  };
});

describe('Extras page in a project', () => {
  it('sends the file name of a new target of a single-file extra', async () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <I18nProvider><ToastProvider><MemoryRouter><ExtrasPage /></MemoryRouter></ToastProvider></I18nProvider>
      </QueryClientProvider>,
    );
    const user = userEvent.setup();

    await user.click(await screen.findByRole('button', { name: 'Add target' }));
    await user.type(screen.getByRole('textbox', { name: 'Folder' }), 'docs/ai');
    await user.type(screen.getByRole('textbox', { name: 'File name' }), 'instructions.md');
    await user.click(screen.getByRole('button', { name: 'Add target' }));

    expect(api.addExtraTarget).toHaveBeenCalledWith('team-rules', { path: 'docs/ai', mode: 'merge', flatten: false, as: 'instructions.md' });
  });
});
