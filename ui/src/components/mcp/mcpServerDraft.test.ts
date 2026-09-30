import { describe, expect, it } from 'vitest';
import { initialServerDraft } from './mcpServerDraft';

describe('MCP server drafts', () => {
  it('starts a new server on the built-in mode, with its option draft under that mode', () => {
    const draft = initialServerDraft(undefined, '', ['pi'], false);
    expect(draft.piExtension).toBe('builtin');
    expect(Object.keys(draft.modeDrafts)).toEqual([draft.piExtension]);
  });

  it('keeps the mode of a server being edited', () => {
    expect(initialServerDraft({ command: 'docs', piExtension: 'pi-mcp-adapter' }, 'docs', ['pi'], false).piExtension).toBe('pi-mcp-adapter');
  });
});
