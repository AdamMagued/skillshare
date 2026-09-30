import type { ReactNode } from 'react';
import CodeEditor from '../CodeEditor';
import { Checkbox, Select } from '../Input';
import { useT } from '../../i18n';
import PiExtensionField from './PiExtensionField';
import { piExposures, type parsePiOptions } from './mcpView';

interface Props {
  mode: string;
  /** Omitted for an off switch, which has no mode to pick. */
  onMode?: (mode: string) => void;
  optionsText: string;
  options: ReturnType<typeof parsePiOptions>;
  optionsError: string;
  onOptions: (text: string) => void;
  keepsOptions: boolean;
  prune: boolean;
  onPrune: (prune: boolean) => void;
  disabled: boolean;
  project: boolean;
  /** Shown between the exposure and the JSON settings, such as the adapter's direct tools. */
  children?: ReactNode;
}

/** Pi's mode, tool exposure, other settings and prune switch, shared by the form and a single pasted server. */
export default function PiSettingsFields({ mode, onMode, optionsText, options, optionsError, onOptions, keepsOptions, prune, onPrune, disabled, project, children }: Props) {
  const t = useT();
  return (
    <>
      {onMode && <PiExtensionField value={mode} onChange={onMode} disabled={disabled} project={project} />}
      {mode === 'builtin' && <div className="ss-fld">
        <Select label={t('mcp.piExposure')} value={String(options.value?.exposure ?? '')} disabled={disabled || Boolean(optionsError)} onChange={(value) => {
          const next = { ...options.value };
          if (value) next.exposure = value; else delete next.exposure;
          onOptions(JSON.stringify(next, null, 2));
        }} options={[{ value: '', label: t('mcp.directToolsUnset'), note: `· ${t('mcp.piExposureUnset')}` }, ...piExposures.map((value) => ({ value, label: value, note: `· ${t(`mcp.piExposure.${value}`)}` }))]} />
        <span className="hp">{t('mcp.piExposureHint')}</span>
      </div>}
      {children}
      {keepsOptions && <div className="ss-fld">
        <label>{t('mcp.piOptions')}</label>
        <CodeEditor value={optionsText} onChange={onOptions} lang="json" placeholder={mode === 'builtin' ? '{\n  "timeout": 120,\n  "toolExposure": {"delete_*": "hidden"}\n}' : '{\n  "excludeTools": ["delete_*"]\n}'} ariaLabel={t('mcp.piOptions')} disabled={disabled} minHeight="96px" />
        {optionsError ? <span className="hp !text-bad">{optionsError}</span> : <span className="hp">{t('mcp.piOptionsHint')}</span>}
        <Checkbox label={t('mcp.piPrune')} checked={prune} onChange={onPrune} disabled={disabled} />
        <span className="hp">{t('mcp.piPruneHint')}</span>
      </div>}
    </>
  );
}
