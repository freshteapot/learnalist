import typescript from '@rollup/plugin-typescript';
import alias from '@rollup/plugin-alias';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import svelte from 'rollup-plugin-svelte';
import { terser } from 'rollup-plugin-terser';
import del from 'rollup-plugin-delete';
import css from 'rollup-plugin-css-only';
import sveltePreprocess from 'svelte-preprocess';
import copy from 'rollup-plugin-copy'


const production = !process.env.ROLLUP_WATCH;
import { getComponentInfo, rollupPluginManifestSync } from "../utils/glue.mjs";

// External and replacement needs to be the full path :(
export default (key, format) => {
    if (format === undefined) {
        format = "umd";
    }

    const componentKey = key;
    const componentInfo = getComponentInfo(componentKey, !production);

    return {
        external: ['shared'],
        input: `src/${componentKey}.js`,
        output: {
            globals: {
                'shared': 'shared',
            },
            sourcemap: !production,
            format: format, // if I want to use globals, this is the way
            name: componentInfo.componentKey,
            dir: componentInfo.localBasePath,
        },
        plugins: [
            alias({
                entries: [
                    { find: '../shared.js', replacement: 'shared' },
                    { find: '../shared.js', replacement: 'shared' },
                    { find: '../../shared.js', replacement: 'shared' },
                    { find: '../../../shared.js', replacement: 'shared' },
                ]
            }),
            del({ targets: componentInfo.rollupDeleteTargets, verbose: true, force: true }),
            typescript(),

            svelte({
                onwarn: (warning, handler) => {
                    const { code, frame } = warning;
                    if (code === "css-unused-selector")
                        return;

                    handler(warning);
                },
                exclude: /\.wc\.svelte$/,
                preprocess: sveltePreprocess({
                    postcss: {
                        configFilePath: "./postcss.config.js",
                    },
                }),
                compilerOptions: {
                    // enable run-time checks when not in production
                    dev: !production,
                    customElement: false,
                },
            }),
            css({ output: `${componentKey}.css` }),

            // Little hack to move the above, due to css not working with relative or outside of folder
            copy({
                targets: componentInfo.rollupCopyTargets,
                verbose: true, force: true,
                hook: 'writeBundle'
            }),

            // TODO Css is not coming thru when customelement includes non-custom element
            // Possible tip https://github.com/sveltejs/svelte/issues/4274.
            svelte({
                include: /\.wc\.svelte$/,
                preprocess: sveltePreprocess({
                    postcss: {
                        configFilePath: "./postcss.config.js",
                    },
                }),
                compilerOptions: {
                    // enable run-time checks when not in production
                    dev: !production,
                    customElement: true,
                },
                onwarn: (warning, handler) => {
                    const { code, frame } = warning;
                    if (code === "css-unused-selector")
                        return;

                    handler(warning);
                },
            }),

            resolve({
                browser: true,
                dedupe: ['svelte']
            }),
            commonjs(),

            /**
             * Minifies JavaScript bundle in production
             */
            production && terser(),

            /**
             * Sync the new filename to hugo, for instant feedback
             */

            rollupPluginManifestSync(componentInfo)
            // TODO how to move from here
        ]
    }
}
