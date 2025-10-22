import adapter from '@sveltejs/adapter-auto';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

const config = {
  // use vitePreprocess from @sveltejs/vite-plugin-svelte
  preprocess: vitePreprocess(),

  kit: {
    // static adapter outputs to /public as we used earlier
    adapter: adapter({
      pages: 'public',
      assets: 'public',
      fallback: null
    }),
    // keep older setting if you want to disable strict origin CSRF checks
    csrf: { checkOrigin: false }
    // NOTE: do NOT add a `vite` property here — Vite config goes in vite.config.ts
  }
};

export default config;
