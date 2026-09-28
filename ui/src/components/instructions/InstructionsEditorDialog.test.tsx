import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { I18nProvider } from '../../i18n';
import { ToastProvider } from '../Toast';
import InstructionsEditorDialog from './InstructionsEditorDialog';

vi.mock('../CodeEditor', () => ({
  default: ({ value, onChange, ariaLabel }: { value: string; onChange: (v: string) => void; ariaLabel: string }) => <textarea aria-label={ariaLabel} value={value} onChange={(e) => onChange(e.target.value)} />,
}));

const renderDialog = (onSave = vi.fn().mockResolvedValue(undefined)) => {
  render(
    <I18nProvider><ToastProvider>
      <InstructionsEditorDialog title="personal" path="/h/personal/AGENTS.md" content={'# Rules\n'} onSave={onSave} onClose={() => {}} />
    </ToastProvider></I18nProvider>,
  );
  return onSave;
};

describe('Instructions editor dialog', () => {
  it('previews the unsaved draft, with HTML comments left out', async () => {
    renderDialog();
    const user = userEvent.setup();

    await user.type(screen.getByRole('textbox', { name: 'personal' }), '<!-- hidden -->{enter}## Draft heading');
    await user.click(screen.getByRole('tab', { name: 'Preview' }));

    expect(screen.getByRole('heading', { name: 'Draft heading' })).toBeInTheDocument();
    expect(screen.queryByText(/hidden/)).not.toBeInTheDocument();
  });

  it('saves the draft with ⌘S while previewing', async () => {
    const onSave = renderDialog();
    const user = userEvent.setup();

    await user.type(screen.getByRole('textbox', { name: 'personal' }), 'more');
    await user.click(screen.getByRole('tab', { name: 'Preview' }));
    fireEvent.keyDown(window, { key: 's', metaKey: true });

    await waitFor(() => expect(onSave).toHaveBeenCalledWith('# Rules\nmore'));
  });
});
