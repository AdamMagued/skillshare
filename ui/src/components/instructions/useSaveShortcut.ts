import { useEffect, useLayoutEffect, useRef } from 'react';
import { claimSaveShortcut } from '../../hooks/useGlobalShortcuts';

/**
 * ⌘S / Ctrl+S calls save instead of opening the browser's save-page dialog,
 * and instead of the dashboard's go-to-Sync shortcut, while enabled. save
 * decides whether there is anything to save.
 */
export function useSaveShortcut(save: () => void, enabled = true) {
  const saveRef = useRef(save);
  useLayoutEffect(() => {
    saveRef.current = save;
  });
  useEffect(() => {
    if (!enabled) return;
    const release = claimSaveShortcut();
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 's') {
        e.preventDefault();
        saveRef.current();
      }
    };
    window.addEventListener('keydown', onKey);
    return () => {
      window.removeEventListener('keydown', onKey);
      release();
    };
  }, [enabled]);
}
