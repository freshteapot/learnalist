import svelte from 'rollup-plugin-svelte';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import typescript from '@rollup/plugin-typescript';
import sveltePreprocess from 'svelte-preprocess';
import css from 'rollup-plugin-css-only';

const production = !process.env.ROLLUP_WATCH;

export default {
	input: 'src/browser-extension/import-play/main.js',
	output: {
		sourcemap: true,
		format: 'iife',
		name: 'app',
		file: 'browser-extension/import-play/bundle.js'
	},
	plugins: [
		typescript(),
		svelte({
			onwarn: (warning, handler) => {
				const { code, frame } = warning;
				if (code === "css-unused-selector")
					return;

				handler(warning);
			},
			preprocess: sveltePreprocess({
				postcss: {
					configFilePath: "./postcss.config.js",
				},
			}),
			compilerOptions: {
				// enable run-time checks when not in production
				dev: !production,
				customElement: false
			},
		}),
		css({ output: 'bundle.css' }),

		resolve({
			browser: true,
			dedupe: ['svelte']
		}),
		commonjs(),
	],
	watch: {
		clearScreen: false
	}
};
