import { Select } from '../Input';
import { useT } from '../../i18n';

export default function PiExtensionField({ value, onChange, disabled = false, project = false }: { value: string; onChange: (value: string) => void; disabled?: boolean; project?: boolean }) {
  const t = useT();
  return <div className="ss-fld">
    <Select label={t('mcp.piExtension')} value={value} onChange={onChange} disabled={disabled} options={[
      { value: '', label: t('mcp.piChoose') },
      { value: 'builtin', label: t('mcp.piBuiltin') },
      { value: 'pi-mcp-adapter', label: 'pi-mcp-adapter' },
      { value: 'pi-mcp-extension', label: 'pi-mcp-extension' },
    ]} />
    <span className="hp">{t(value === 'builtin' ? 'mcp.piBuiltinHelp' : 'mcp.piPrerequisite')}</span>
    {value === 'builtin' && <span className="hp">{t(project ? 'mcp.piProjectHint' : 'mcp.piGlobalHint')}</span>}
    {value && value !== 'builtin' && <><code className="text-xs">pi install npm:{value}</code><span className="hp">{t(value === 'pi-mcp-adapter' ? 'mcp.piAdapterHelp' : 'mcp.piExtensionHelp')}</span></>}
    <span className="flex flex-wrap gap-3 text-xs">
      <a className="text-link underline" href="https://github.com/earendil-works/pi/blob/v0.99.0/packages/coding-agent/docs/mcp.md" target="_blank" rel="noopener noreferrer">Pi MCP · {t('plugins.officialDocs')}</a>
      {['pi-mcp-adapter', 'pi-mcp-extension'].map((pkg) => <a key={pkg} className="text-link underline" href={`https://pi.dev/packages/${pkg}`} target="_blank" rel="noopener noreferrer">{pkg} · {t('plugins.officialDocs')}</a>)}
    </span>
  </div>;
}
