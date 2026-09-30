import { describe, expect, it } from 'vitest';
import { initialServerDraft } from './mcpServerDraft';

describe('MCP server drafts', () => {
  it('initializes the option draft under the active mode when the Pi default is empty', () => {
    const draft = initialServerDraft(undefined, '', ['pi'], '', false);
    expect(draft.piExtension).toBe('builtin');
    expect(Object.keys(draft.modeDrafts)).toEqual([draft.piExtension]);
  });
});
