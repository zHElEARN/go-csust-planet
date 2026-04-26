import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

const backendOrigin = process.env.ADMIN_API_PROXY_TARGET ?? 'http://127.0.0.1:7241';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		proxy: {
			'/v1': {
				target: backendOrigin,
				changeOrigin: true
			}
		}
	}
});
