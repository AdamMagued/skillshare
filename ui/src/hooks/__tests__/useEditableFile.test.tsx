import type { ReactNode } from 'react';
import { act, renderHook, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ToastProvider } from '../../components/Toast';
import { I18nProvider } from '../../i18n';
import { useEditableFile } from '../useEditableFile';

type Data = { raw?: string } | undefined;

const wrapper = ({ children }: { children: ReactNode }) => (
  <I18nProvider><ToastProvider>{children}</ToastProvider></I18nProvider>
);

function renderFile(data: Data, save: (value: string) => Promise<string | void> = async () => {}, skipEmpty = false) {
  return renderHook(({ data }: { data: Data }) => useEditableFile(data, { save, skipEmpty }), { wrapper, initialProps: { data } });
}

describe('useEditableFile', () => {
  it('loads the fetched text', () => {
    const { result } = renderFile({ raw: 'a' });

    expect(result.current.value).toBe('a');
  });

  it('is dirty only while the edit differs from the fetched text', () => {
    const { result } = renderFile({ raw: 'a' });

    act(() => { result.current.change('b'); });
    const edited = result.current.dirty;
    act(() => { result.current.change('a'); });

    expect([edited, result.current.dirty]).toEqual([true, false]);
  });

  it('reset restores the fetched text', () => {
    const { result } = renderFile({ raw: 'a' });

    act(() => { result.current.change('b'); });
    act(() => { result.current.reset(); });

    expect([result.current.value, result.current.dirty]).toEqual(['a', false]);
  });

  it('shows the text the save wrote and clears dirty', async () => {
    const { result } = renderFile({ raw: 'a' }, async (value) => `${value}!`);

    act(() => { result.current.change('b'); });
    await act(() => result.current.save());

    expect([result.current.value, result.current.dirty]).toEqual(['b!', false]);
  });

  it('keeps the edit dirty and toasts when the save fails', async () => {
    const { result } = renderFile({ raw: 'a' }, vi.fn().mockRejectedValue(new Error('disk full')));

    act(() => { result.current.change('b'); });
    await act(() => result.current.save());

    expect([result.current.dirty, screen.queryByText('disk full') !== null]).toEqual([true, true]);
  });

  it('with skipEmpty, keeps the editor when the fetched text is empty', () => {
    const { result, rerender } = renderFile({ raw: 'a' }, undefined, true);

    rerender({ data: { raw: '' } });

    expect(result.current.value).toBe('a');
  });
});
