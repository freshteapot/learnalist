import svelte from 'rollup-plugin-svelte';
import resolve from 'rollup-plugin-node-resolve';
import commonjs from 'rollup-plugin-commonjs';
import livereload from 'rollup-plugin-livereload';
import { terser } from 'rollup-plugin-terser';

const production = !process.env.ROLLUP_WATCH;

export default {
	input: 'src/user.js',
	output: {
		sourcemap: true,
		format: 'iife',
		name: 'app',
		file: 'public/user.js'
	},
	plugins: [
		svelte({
			// enable run-time checks when not in production
			dev: !production,
			customElement : true,
			// we'll extract any component CSS out into
			// a separate file  better for performance
			css: css => {
				css.write('public/user.css');
			}
		}),
		resolve({
			browser: true,
			dedupe: importee => importee === 'svelte' || importee.startsWith('svelte/')
		}),
		commonjs(),
		!production && livereload('public'),
		production && terser()
	],
	watch: {
		clearScreen: false
	}
};
