import { sveltekit } from '@sveltejs/kit/vite';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import adapter from '@sveltejs/adapter-static';
import { defineConfig } from 'vite';

export default defineConfig(({ command }) => ({
	plugins: [
		sveltekit({
			adapter: adapter({
				pages: 'build',
				assets: 'build',
				fallback: undefined,
				precompress: false,
				strict: true,
			}),
			preprocess: [vitePreprocess({})],
			csp: {
				mode: 'hash',
				directives: {
					'default-src': ['self'],
					'script-src': ['self'],
					'style-src': ['self', 'unsafe-inline'],
					'img-src': ['self', 'data:'],
					'connect-src': command === 'serve' ? ['self', 'http://localhost:3000'] : ['self'],
					'object-src': ['none'],
					'base-uri': ['none'],
					'form-action': ['none'],
				},
			},
		}),
	],

	css: {
		preprocessorOptions: {
			scss: {
				additionalData: '@use "/src/variables.scss" as *;',
				// Bootstrap 5 still relies on @import and legacy Sass functions
				quietDeps: true,
				silenceDeprecations: ['import', 'global-builtin', 'color-functions', 'if-function'],
			},
		},
	},
}));
