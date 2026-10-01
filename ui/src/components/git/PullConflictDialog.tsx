import { useState } from 'react';
import type { GitConflictVersion, GitPullConflict, GitPullResolution } from '../../api/types/git';
import { useT } from '../../i18n';
import Button from '../Button';
import CodeView from '../CodeView';
import DialogShell from '../DialogShell';

interface Props {
  conflict: GitPullConflict;
  onCancel: () => void;
  onConfirm: (resolution: GitPullResolution) => void;
}

export default function PullConflictDialog({ conflict, onCancel, onConfirm }: Props) {
  const t = useT();
  const [choices, setChoices] = useState<GitPullResolution['choices']>({});
  const title = t('gitSync.conflict.title');
  const complete = conflict.files.every((file) => choices[file.path]);
  const preview = (version: GitConflictVersion, path: string) => version.deleted
    ? <div className="ss-note warn">{t('gitSync.conflict.deleted')}</div>
    : version.noPreview
      ? <div className="ss-note inf">{t('gitSync.conflict.noPreview')}</div>
      : <CodeView content={version.content} lang={path} className="max-h-52" />;

  return (
    <DialogShell open onClose={onCancel} maxWidth="5xl" padding="none" ariaLabel={title}>
      <div className="dh"><h2 className="ss-h2">{title}</h2></div>
      <div className="db flex flex-col gap-5">
        <div className="ss-note warn">{t('gitSync.conflict.hint')}</div>
        {conflict.files.map((file) => (
          <section key={file.path} className="flex min-w-0 flex-col gap-3">
            <h3 className="break-all font-mono text-[13px] font-semibold">{file.path}</h3>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              {(['local', 'remote'] as const).map((side) => (
                <div key={side} className="flex min-w-0 flex-col gap-2">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-[13px] font-semibold">{t(`gitSync.conflict.${side}`)}</span>
                    <Button size="sm" variant={choices[file.path] === side ? 'primary' : 'secondary'} aria-pressed={choices[file.path] === side} aria-label={t(`gitSync.conflict.choose.${side}`, { path: file.path })} onClick={() => setChoices((previous) => ({ ...previous, [file.path]: side }))}>
                      {t(choices[file.path] === side ? 'gitSync.conflict.selected' : 'gitSync.conflict.choose')}
                    </Button>
                  </div>
                  {preview(file[side], file.path)}
                </div>
              ))}
            </div>
          </section>
        ))}
      </div>
      <div className="df flex-wrap">
        <span className="flex-1 text-[13px] text-ink-2">{t('gitSync.conflict.progress', { chosen: Object.keys(choices).length, total: conflict.files.length })}</span>
        <Button variant="ghost" onClick={onCancel}>{t('common.cancel')}</Button>
        <Button variant="primary" disabled={!complete} onClick={() => onConfirm({ localHash: conflict.localHash, remoteHash: conflict.remoteHash, choices })}>{t('gitSync.conflict.apply')}</Button>
      </div>
    </DialogShell>
  );
}
