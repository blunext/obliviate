import {sveltekit} from "@sveltejs/kit/vite";
import {defineConfig} from "vite";

export default defineConfig({
// Disable code splitting
//     build: {
//         rollupOptions: {
//             output: {
//                 manualChunks: () => 'app',
//             },
//         },
//     },
    plugins: [sveltekit()],
    css: {
        preprocessorOptions: {
            scss: {
                additionalData: '@use "src/variables.scss" as *;',
            },
        },
    },
});
