import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { I18nProvider } from '../i18n';
import { ToastProvider } from '../components/Toast';
import { api } from '../api/client';
import ExtrasPage from './ExtrasPage';

vi.mock('../context/AppContext', () => ({ useAppContext: () => ({ isProjectMode: true }) }));
vi.mock('../api/client', async (load) => {
  const actual = await load<typeof import('../api/client')>();
  return {
    ...actual,
    api: {
      ...actual.api,
      listExtras: vi.fn().mockResolvedValue({
        extras: [
          { name: 'team-rules', file: 'AGENTS.md', source_dir: '/p/.skillshare/extras/team-rules', source_type: 'per-extra', file_count: 1, source_exists: true,
            targets: [{ path: '.', mode: 'symlink', flatten: false, status: 'synced' }] },
          { name: 'rules', source_dir: '/p/.skillshare/extras/rules', source_type: 'per-extra', file_count: 2, source_exists: true,
            targets: [{ path: '.claude/rules', mode: 'merge', flatten: false, status: 'synced' }] },
          { name: 'conventions', file: 'CONVENTIONS.md', source_dir: '/p/.skillshare/extras/conventions', source_type: 'per-extra', file_count: 1, source_exists: true,
            targets: [{ path: '.cursor', mode: 'copy', flatten: false, as: 'rules.md', status: 'synced' }] },
        ],
      }),
      createExtra: vi.fn().mockResolvedValue({ success: true }),
      listExtraExtensions: vi.fn().mockResolvedValue({ extensions: [] }),
      availableTargets: vi.fn().mockResolvedValue({ targets: [] }),
      getOverview: vi.fn().mockResolvedValue({}),
    },
  };
});

const renderPage = () => render(
  <QueryClientProvider client={new QueryClient()}>
    <I18nProvider><ToastProvider><MemoryRouter><ExtrasPage /></MemoryRouter></ToastProvider></I18nProvider>
  </QueryClientProvider>,
);

describe('Extras page in a project', () => {
  it('lists folder extras on the first tab', async () => {
    renderPage();
    expect(await screen.findByText('rules')).toBeInTheDocument();
  });

  // Shared AGENTS.md files are on the AGENTS.md tab.
  it('leaves AGENTS.md extras off the first tab', async () => {
    renderPage();
    await screen.findByText('rules');
    expect(screen.queryByText('team-rules')).not.toBeInTheDocument();
  });

  it('shows other single-file extras with the full target file path', async () => {
    renderPage();
    expect(await screen.findByText('conventions')).toBeInTheDocument();
    expect(screen.getByText('.cursor/rules.md')).toBeInTheDocument();
  });

  it('creates a single-file extra with its file and target file name', async () => {
    const user = userEvent.setup();
    renderPage();
    await user.click(await screen.findByRole('button', { name: 'Add extra' }));
    const dialog = screen.getByRole('dialog');
    await user.type(within(dialog).getByLabelText('Name'), 'notes');
    await user.click(within(dialog).getByRole('radio', { name: 'Single file' }));
    const [fileInput, asInput] = within(dialog).getAllByRole('textbox', { name: 'File name' });
    await user.type(fileInput, 'NOTES.md');
    await user.type(within(dialog).getByRole('textbox', { name: 'Folder' }), '.claude');
    await user.type(asInput, 'CLAUDE-notes.md');
    await user.click(within(dialog).getByRole('button', { name: 'Create' }));

    expect(api.createExtra).toHaveBeenCalledWith({
      name: 'notes',
      file: 'NOTES.md',
      targets: [{ path: '.claude', mode: 'merge', as: 'CLAUDE-notes.md' }],
    });
  });
});
