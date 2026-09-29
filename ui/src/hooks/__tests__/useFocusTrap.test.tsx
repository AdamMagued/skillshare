import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { useFocusTrap } from '../useFocusTrap';

function Trap({ autoFocusInput }: { autoFocusInput: boolean }) {
  const ref = useFocusTrap(true);
  return (
    <div ref={ref}>
      <button type="button">Close</button>
      <input aria-label="Name" autoFocus={autoFocusInput} />
    </div>
  );
}

describe('useFocusTrap', () => {
  it('keeps the focus a field inside already took', () => {
    render(<Trap autoFocusInput />);
    expect(screen.getByLabelText('Name')).toHaveFocus();
  });

  it('focuses the first control otherwise', () => {
    render(<Trap autoFocusInput={false} />);
    expect(screen.getByRole('button', { name: 'Close' })).toHaveFocus();
  });
});
