import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import MarkdownView from './MarkdownView';

const view = (md: string, breaks?: boolean) => render(<MarkdownView breaks={breaks}>{md}</MarkdownView>).container;

describe('MarkdownView', () => {
  it('renders HTML such as <details> but drops scripts and event handlers', () => {
    const html = view('<details><summary>More</summary>\n\nhidden\n\n</details>\n\n<script>alert(1)</script>\n<img src="x" onerror="alert(1)">').innerHTML;
    expect([html.includes('<details>'), html.includes('<script'), html.includes('onerror')]).toEqual([true, false, false]);
  });

  it('shows a GitHub alert and a ::: container as the same alert', () => {
    const alerts = view('> [!WARNING]\n> Careful.\n\n::: warning\n*dragons*\n:::').querySelectorAll('blockquote.alert.warning');
    expect([...alerts].map((a) => a.textContent!.replace(/\s+/g, ''))).toEqual(['WarningCareful.', 'Warningdragons']);
  });

  it('reads ==mark==, ++ins++, ^sup^ and ~sub~ only when the marks hug the text', () => {
    const c = view('==a== ++b++ 19^th^ H~2~O, but C++ and C++, a == b, x^2 + y^2');
    expect(['mark', 'ins', 'sup', 'sub'].map((t) => [...c.querySelectorAll(t)].map((e) => e.textContent).join())).toEqual(['a', 'b', 'th', '2']);
  });

  it('treats a leading --- block that is not YAML as text, not frontmatter', () => {
    const c = view('---\n__Ad :)__\n\n- a\n\n---\n\n# Title');
    expect([c.querySelector('.ss-tbl.fm'), c.querySelector('h1')?.textContent]).toEqual([null, 'Title']);
  });

  it('links each footnote reference to its note', () => {
    const c = view('Text[^1].\n\n[^1]: Note.');
    const href = c.querySelector('a[data-footnote-ref]')!.getAttribute('href')!;
    expect(c.querySelector(href)).not.toBeNull();
  });

  it('keeps single line breaks only when asked', () => {
    expect([view('one\ntwo').querySelectorAll('br').length, view('one\ntwo', true).querySelectorAll('br').length]).toEqual([0, 1]);
  });
});
