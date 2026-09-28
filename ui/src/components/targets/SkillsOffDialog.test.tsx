import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { api } from '../../api/client';
import type { Target } from '../../api/client';
import { I18nProvider } from '../../i18n';
import SkillsOffDialog from './SkillsOffDialog';

vi.mock('../../api/client', async (load) => ({ ...await load<typeof import('../../api/client')>(), api: { skillsOffPreview: vi.fn(), updateTarget: vi.fn() } }));

const gemini = { name: 'gemini', path: '/home/me/.gemini/skills' } as Target;
const universal = { name: 'universal', path: '/home/me/.agents/skills', linkedCount: 9 } as Target;
const view = (props: Partial<Parameters<typeof SkillsOffDialog>[0]> = {}) => {
  const onStopped = vi.fn();
  render(
    <QueryClientProvider client={new QueryClient()}>
      <I18nProvider><SkillsOffDialog target={gemini} managed={['MCP', 'GEMINI.md']} onClose={vi.fn()} onStopped={onStopped} {...props} /></I18nProvider>
    </QueryClientProvider>,
  );
  return onStopped;
};

describe('Skills off dialog', () => {
  beforeEach(() => vi.clearAllMocks());

  it('lists what stopping removes and keeps, then stops and reports the removed links', async () => {
    const user = userEvent.setup();
    vi.mocked(api.skillsOffPreview).mockResolvedValue({ remove: ['archify', 'hub'], keep: ['my-notes'] });
    vi.mocked(api.updateTarget).mockResolvedValue({ success: true, detach: { removed: ['archify', 'hub'], kept: ['my-notes'] } });
    const onStopped = view({ readFrom: universal });
    expect(await screen.findByText(/Remove the 2 links skillshare made in/)).toBeInTheDocument();
    expect(screen.getByText('Keep 1 item of your own: my-notes')).toBeInTheDocument();
    expect(screen.getByText('MCP and GEMINI.md stay managed as usual')).toBeInTheDocument();
    expect(screen.getByText(/gemini also reads universal’s .*9 skills there/)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Stop syncing' }));
    await waitFor(() => expect(onStopped).toHaveBeenCalledWith(2));
    expect(api.updateTarget).toHaveBeenCalledWith('gemini', { skills_enabled: false });
  });

  it('says nothing is removed when an enabled target shares the folder', async () => {
    vi.mocked(api.skillsOffPreview).mockResolvedValue({ remove: [], keep: [], sharedWith: 'universal' });
    view();
    expect(await screen.findByText(/is also the skills folder of universal, so nothing is removed/)).toBeInTheDocument();
  });

  it('warns who loses the skills when others read the folder being turned off, counting past a few', async () => {
    vi.mocked(api.skillsOffPreview).mockResolvedValue({ remove: ['archify'], keep: [] });
    view({ target: universal, managed: ['AGENTS.md'], readers: { off: ['gemini'], local: ['codex', 'copilot', 'droid', 'grok', 'zed', 'amp', 'cline'], on: ['pi'] } });
    expect(await screen.findByText(/they read only this folder/)).toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'gemini' })).toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'codex, copilot, droid, grok, zed, amp, cline' })).toBeInTheDocument();
    expect(screen.getByText('and 2 more')).toBeInTheDocument();
  });
});
