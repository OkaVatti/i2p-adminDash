<script>
  import { onDestroy, onMount } from 'svelte';
  import { theme } from '$lib/stores';
  let current = 'Default';
  const themeMap = {
    Default: null,
    Light: '/themes/light.css',
    Dark: '/themes/dark.css',
    LightPlus: '/themes/lightplus.css',
    DarkPlus: '/themes/darkplus.css',
    Midnight: '/themes/midnight.css',
    OLED: '/themes/oled.css',
    Platinum: '/themes/platinum.css'
  };

  const applyTheme = (t) => {
    const href = themeMap[t] || null;
    // remove existing theme link
    let el = document.getElementById('i2p-theme');
    if (el) el.remove();
    if (href) {
      el = document.createElement('link');
      el.rel = 'stylesheet';
      el.id = 'i2p-theme';
      el.href = href;
      document.head.appendChild(el);
      document.documentElement.classList.remove('oled', 'midnight');
    } else {
      // Default: no extra stylesheet (use I2P default later)
    }
  };

  const unsub = theme.subscribe((v) => {
    current = v;
    if (typeof document !== 'undefined') applyTheme(v);
  });

  onMount(() => {
    // apply initial theme
    applyTheme(current);
  });

  // cleanup
  onDestroy(() => {
    unsub();
  });
</script>

<slot />
