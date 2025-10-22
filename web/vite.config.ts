import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
	],
	server: {
		// allow serving files from project root (same as previous server.fs.allow)
		fs: {
			allow: ['.']
		}
	}

});
