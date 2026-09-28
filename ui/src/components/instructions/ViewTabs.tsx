import MarkdownView from '../MarkdownView';
import { useT } from '../../i18n';
import { previewParts } from './instructionsView';

/**
 * Underline tabs that switch how a box shows an instruction file (Edit /
 * Preview, or Source / Preview when it is read-only). className replaces the
 * strip's own divider and padding, for a header that already has them.
 */
export function ViewTabs<V extends string>({ view, views: given, onChange, className = 'border-b border-line-soft px-4' }: {
  view: V;
  /** Defaults to Edit / Preview (values edit and preview). */
  views?: { value: V; label: string }[];
  onChange: (view: V) => void;
  className?: string;
}) {
  const t = useT();
  const views = given ?? ([
    { value: 'edit', label: t('instructions.target.view.edit') },
    { value: 'preview', label: t('instructions.target.view.preview') },
  ] as { value: V; label: string }[]);
  return (
    <div role="tablist" aria-label={t('instructions.target.view.label')} className={`flex gap-5 ${className}`}
      onKeyDown={(e) => {
        if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return;
        e.preventDefault();
        const i = views.findIndex((v) => v.value === view);
        const next = views[(i + (e.key === 'ArrowRight' ? 1 : views.length - 1)) % views.length].value;
        onChange(next);
        e.currentTarget.querySelector<HTMLElement>(`[data-view="${next}"]`)?.focus();
      }}>
      {views.map((v) => (
        <button key={v.value} type="button" role="tab" data-view={v.value} aria-selected={view === v.value} tabIndex={view === v.value ? 0 : -1} onClick={() => onChange(v.value)}
          className={`-mb-px h-9 border-b-2 text-[12.5px] ${view === v.value ? 'border-ink font-semibold text-ink' : 'border-transparent text-ink-2 hover:text-ink'}`}>
          {v.label}
        </button>
      ))}
    </div>
  );
}

/**
 * An instruction file rendered as Markdown: HTML comments are dropped and the
 * managed import block shows as the shared files it imports (names: known
 * shared file names).
 */
export function InstructionsPreview({ content, names, className = 'min-h-0 flex-1 overflow-auto' }: { content: string; names: string[]; className?: string }) {
  const t = useT();
  const { before, imports, after } = previewParts(content, names);
  const empty = !before.trim() && imports.length === 0 && !after.trim();
  return (
    <div className={`flex flex-col gap-3 px-6 py-5 text-[13.5px] ${className}`} style={{ fontFamily: 'var(--f)' }}
      role="tabpanel" aria-label={t('instructions.target.view.preview')}>
      {before.trim() && <MarkdownView>{before}</MarkdownView>}
      {imports.length > 0 && <p className="font-mono text-[12px] text-ink-3">{imports.map((n) => `@import ${n}`).join(' · ')}</p>}
      {after.trim() && <MarkdownView>{after}</MarkdownView>}
      {empty && <p className="text-[13px] text-ink-3">{t('instructions.target.previewEmpty')}</p>}
    </div>
  );
}
