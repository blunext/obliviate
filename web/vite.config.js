import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],

	css: {
		preprocessorOptions: {
			scss: {
				additionalData: '@use "$lib/../variables.scss" as *;',
				// Bootstrap 5 still relies on @import and legacy Sass functions
				quietDeps: true,
				silenceDeprecations: ['import', 'global-builtin', 'color-functions', 'if-function'],
			},
		},
	},
});
