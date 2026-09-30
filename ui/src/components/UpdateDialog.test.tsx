import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { I18nProvider } from '../i18n';
import UpdateDialog from './UpdateDialog';
import { api } from '../api/client';

vi.mock('../api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/client')>();
  return {
    ...actual,
    api: {
      ...actual.api,
      getVersionCheck: vi.fn().mockResolvedValue({
        cliVersion: '0.21.12',
        cliLatest: '0.21.13',
        cliUpdateAvailable: true,
        skillVersion: '0.21.12',
        skillLatest: '0.21.13',
        skillUpdateAvailable: true,
      }),
      upgradeApp: vi.fn().mockResolvedValue({}),
      restartApp: vi.fn().mockResolvedValue({ ok: true, restarting: true }),
      health: vi.fn(() => new Promise(() => {})),
    },
  };
});

function renderDialog() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <I18nProvider>
        <UpdateDialog />
      </I18nProvider>
    </QueryClientProvider>,
  );
}

describe('UpdateDialog', () => {
  it('stays closed on a dev build even when an update is reported', async () => {
    vi.mocked(api.getVersionCheck).mockResolvedValueOnce({
      cliVersion: 'dev',
      cliLatest: 'dev-ui-flow',
      cliUpdateAvailable: true,
      cliDevMode: true,
      skillVersion: '0.20.25',
      skillLatest: '0.22.0',
      skillUpdateAvailable: true,
    });
    renderDialog();

    await waitFor(() => expect(api.getVersionCheck).toHaveBeenCalled());
    await act(() => new Promise((resolve) => setTimeout(resolve, 50)));
    expect(screen.queryByRole('button', { name: /Update now/ })).toBeNull();
  });

  it('keeps the UI the upgrade just downloaded when restarting', async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <I18nProvider>
          <UpdateDialog />
        </I18nProvider>
      </QueryClientProvider>,
    );

    await userEvent.click(await screen.findByRole('button', { name: /Update now/ }));

    await waitFor(() => expect(api.restartApp).toHaveBeenCalledWith({ clearCache: false }));
  });
});
