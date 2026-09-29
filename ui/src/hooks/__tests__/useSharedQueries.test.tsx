import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { api } from '../../api/client';
import type { Overview } from '../../api/client';
import { mcpApi } from '../../api/mcp';
import { useMcpQuery, useOverviewQuery } from '../useSharedQueries';

vi.mock('../../api/client', async (load) => {
  const actual = await load<typeof import('../../api/client')>();
  return { ...actual, api: { ...actual.api, getOverview: vi.fn() } };
});
vi.mock('../../api/mcp', async (load) => ({ ...await load<typeof import('../../api/mcp')>(), mcpApi: { list: vi.fn() } }));

function wrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: ReactNode }) => <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

describe('shared queries', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.getOverview).mockResolvedValue({ version: '1.0.0' } as Overview);
  });

  it('serves every overview reader from one request', async () => {
    const { result } = renderHook(() => [useOverviewQuery(), useOverviewQuery()], { wrapper: wrapper() });

    await waitFor(() => expect(result.current.every((q) => q.isSuccess)).toBe(true));

    expect(api.getOverview).toHaveBeenCalledOnce();
  });

  it('keeps a caller option such as enabled', () => {
    renderHook(() => useMcpQuery({ enabled: false }), { wrapper: wrapper() });

    expect(mcpApi.list).not.toHaveBeenCalled();
  });
});
