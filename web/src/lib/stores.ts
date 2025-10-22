// web/src/lib/stores.ts
import { writable, type Writable } from 'svelte/store';

function createThemeStore(): Writable<string> {
  let initial = 'Default';
  try {
    const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('i2p_theme') : null;
    if (stored) initial = stored;
  } catch (e) {
    // ignore (SSR)
  }
  const store = writable<string>(initial);
  store.subscribe((v) => {
    try {
      if (typeof localStorage !== 'undefined') localStorage.setItem('i2p_theme', v);
    } catch (e) {
      // ignore
    }
  });
  return store;
}

export const token = writable<string | null>(null);
export const theme = createThemeStore();
