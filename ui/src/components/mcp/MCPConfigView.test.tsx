import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import { mcpApi, type MCPMutation } from '../../api/mcp';
import { I18nProvider } from '../../i18n';
import MCPConfigView from './MCPConfigView';

vi.mock('../CopyButton', () => ({ default: () => null }));
vi.mock('../../api/mcp', async (load) => ({ ...await load<typeof import('../../api/mcp')>(), mcpApi: { render: vi.fn() } }));

it('keeps the previous native preview while a changed draft is rendering', async () => {
  const client = new QueryClient();
  const initial = { rendered: [{ target: 'pi', path: '/pi/mcp.json', content: 'previous native preview' }] };
  let resolve!: (value: typeof initial) => void;
  vi.mocked(mcpApi.render).mockResolvedValueOnce(initial).mockImplementationOnce(() => new Promise((done) => { resolve = done; }));
  const view = (command: string) => {
    const mutation: MCPMutation = { name: 'docs', server: { command, piExtension: 'builtin', targets: ['pi'] } };
    return <QueryClientProvider client={client}><I18nProvider><MCPConfigView mutation={mutation} /></I18nProvider></QueryClientProvider>;
  };
  const result = render(view('old'));
  expect(await screen.findByText('previous native preview')).toBeInTheDocument();
  result.rerender(view('new'));
  await waitFor(() => expect(mcpApi.render).toHaveBeenCalledTimes(2));
  expect(screen.getByText('previous native preview')).toBeInTheDocument();
  resolve({ rendered: [{ target: 'pi', path: '/pi/mcp.json', content: 'updated native preview' }] });
  expect(await screen.findByText('updated native preview')).toBeInTheDocument();
});
