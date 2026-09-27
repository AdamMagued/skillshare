import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { api } from '../../api/client';
import type { TargetInstructions as Data } from '../../api/client';
import { I18nProvider } from '../../i18n';
import { ToastProvider } from '../Toast';
import TargetInstructions from './TargetInstructions';

vi.mock('../CodeEditor', () => ({
  default: ({ value, onChange, ariaLabel }: { value: string; onChange: (v: string) => void; ariaLabel: string }) => <textarea aria-label={ariaLabel} value={value} onChange={(e) => onChange(e.target.value)} />,
}));
vi.mock('../../api/client', async (load) => {
  const actual = await load<typeof import('../../api/client')>();
  return { ...actual, api: { ...actual.api, getTargetInstructions: vi.fn() } };
});

const file = (target: string, path: string, extra: Partial<Data> = {}): Data => ({
  target, project: false, supported: true, custom: false, path, exists: true, content: `${target} file\n`, size: 10, import: false,
  read_order: [{ path, kind: 'main', exists: true, read: true }], import_lines: [], shared: [], convert: [], riders: [], read_by: [], ...extra,
});

// The tool is picked from a dropdown: open it, then choose the option.
const pickTool = async (user: ReturnType<typeof userEvent.setup>, name: RegExp) => {
  await user.click(await screen.findByRole('combobox', { name: 'Instruction files' }));
  await user.click(await screen.findByRole('option', { name }));
};

const renderUniversal = () => {
  vi.mocked(api.getTargetInstructions).mockImplementation(async (name) => name === 'codex'
    ? file('codex', '~/.codex/AGENTS.md', { rider_of: 'universal' })
    : file('universal', '~/.agents/AGENTS.md', { riders: [{ name: 'codex', path: '~/.codex/AGENTS.md', exists: true }], read_by: ['cline', 'warp'] }));
  render(
    <MemoryRouter initialEntries={['/targets/universal?tab=instructions']}>
      <QueryClientProvider client={new QueryClient()}>
        <I18nProvider><ToastProvider><TargetInstructions name="universal" skillsPath="~/.agents/skills" /></ToastProvider></I18nProvider>
      </QueryClientProvider>
    </MemoryRouter>,
  );
};

describe('Target instructions tab', () => {
  // jsdom has no scrollIntoView, which the dropdown calls on its focused option.
  beforeEach(() => {
    HTMLElement.prototype.scrollIntoView = vi.fn();
  });

  // Codex reads skills from ~/.agents/skills but its own ~/.codex/AGENTS.md, so universal offers it.
  it('switches from universal to the file of a tool that reads its skills', async () => {
    renderUniversal();
    const user = userEvent.setup();

    await pickTool(user, /Codex/);
    await waitFor(() => expect(api.getTargetInstructions).toHaveBeenCalledWith('codex'));
  });

  it('asks before switching away from an unsaved edit and keeps it on cancel', async () => {
    renderUniversal();
    const user = userEvent.setup();

    await user.type(await screen.findByRole('textbox', { name: 'AGENTS.md' }), 'draft');
    await pickTool(user, /Codex/);
    await user.click(await screen.findByRole('button', { name: 'Cancel' }));

    expect(screen.getByRole('textbox', { name: 'AGENTS.md' })).toHaveValue('universal file\ndraft');
  });
});
