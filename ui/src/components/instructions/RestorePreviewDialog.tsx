import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { X } from 'lucide-react';
import { api } from '../../api/client';
import Button from '../Button';
import DialogShell from '../DialogShell';
import Spinner from '../Spinner';
import { formatDateTime, useI18n, useT } from '../../i18n';
import { shortenHome } from '../../lib/paths';
import { queryKeys } from '../../lib/queryKeys';
import { formatSize, lineCount, lineDiff } from './instructionsView';

/** Restore one target: shows what its file goes back to before taking the shared file off it. */
export default function RestorePreviewDialog({ name, target, label, busy, onConfirm, onClose }: {
  name: string;
  target: string;
  /** The target as the row names it (Codex for a tool that is not a target). */
  label: string;
  busy: boolean;
  onConfirm: () => void;
  onClose: () => void;
}) {
  const t = useT();
  const { locale } = useI18n();
  const [tab, setTab] = useState<'content' | 'diff'>('content');
  const { data, error } = useQuery({ queryKey: queryKeys.instructions.restorePreview(name, target), queryFn: () => api.getSharedRestorePreview(name, target), gcTime: 0 });
  const title = t('instructions.restorePreview.title', { name, target: label });
  const lines = data ? data.content.replace(/\n$/, '').split('\n') : [];

  return (
    <DialogShell open onClose={onClose} padding="none" preventClose={busy} ariaLabel={title} className="!max-w-[720px]">
      <div className="dh">
        <div className="flex min-w-0 flex-col gap-1">
          <h2 className="ss-h2">{title}</h2>
          {data && (
            <p className="text-[13px] text-ink-2">
              {data.recorded_at
                ? t('instructions.restorePreview.when', { path: shortenHome(data.path), name, time: formatDateTime(data.recorded_at, locale, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }) })
                : t('instructions.restorePreview.whenUnknown', { path: shortenHome(data.path), name })}
            </p>
          )}
        </div>
        <button type="button" className="ss-ib" aria-label={t('common.close')} onClick={onClose} disabled={busy}><X size={16} /></button>
      </div>
      <div className="db flex flex-col gap-3">
        {error ? (
          <div className="ss-note bad"><span className="flex-1">{error.message}</span></div>
        ) : !data ? (
          <Spinner className="self-center" />
        ) : data.kind === 'delete' ? (
          <p className="text-[13px]">{t('instructions.restorePreview.delete')}</p>
        ) : data.kind === 'link' ? (
          <p className="text-[13px]">{t('instructions.restorePreview.link')} <span className="font-mono">{shortenHome(data.link_to ?? '')}</span></p>
        ) : (
          <>
            <div className="ss-tabs" role="tablist">
              {(['content', 'diff'] as const).map((k) => (
                <button key={k} type="button" role="tab" aria-selected={tab === k} className={tab === k ? 'on' : ''} onClick={() => setTab(k)}>
                  {t(`instructions.restorePreview.tab.${k}`)}
                </button>
              ))}
            </div>
            <div className="ss-list !shadow-none" role="tabpanel">
              <div className="ss-lh !normal-case">
                <span className="flex-1 text-[12.5px]">
                  {tab === 'content'
                    ? t('instructions.preview.stats', { lines: lineCount(data.content), size: formatSize(new TextEncoder().encode(data.content).length) })
                    : shortenHome(data.path)}
                </span>
              </div>
              {tab === 'content' ? (
                <pre className="ss-code !max-h-[320px] !overflow-auto !rounded-none !border-0 !whitespace-pre-wrap" style={{ overflowWrap: 'anywhere' }}>
                  {/* Long lines wrap beside their number. */}
                  {lines.map((l, i) => <span key={i} className="flex"><span className="ln w-6 shrink-0 text-right">{i + 1}</span><span className="min-w-0 flex-1">{l || ' '}</span></span>)}
                </pre>
              ) : (
                <pre className="ss-code !max-h-[320px] !overflow-auto !rounded-none !border-0 !whitespace-pre-wrap" style={{ overflowWrap: 'anywhere' }}>
                  {lineDiff(data.current, data.content).map((l, i) => (
                    <span key={i} className={l.kind === 'same' ? 'block' : l.kind}>{l.kind === 'add' ? '+ ' : l.kind === 'del' ? '− ' : '  '}{l.text || ' '}</span>
                  ))}
                </pre>
              )}
            </div>
          </>
        )}
        {data?.drift && <p className="text-[12.5px] text-ink-2">{t('instructions.restorePreview.drift', { target: label })}</p>}
      </div>
      <div className="df">
        <Button variant="ghost" onClick={onClose} disabled={busy}>{t('common.cancel')}</Button>
        <Button variant="primary" onClick={onConfirm} loading={busy} disabled={!data}>{t('instructions.detail.restore')}</Button>
      </div>
    </DialogShell>
  );
}
