import typescript from '@rollup/plugin-typescript';
import alias from '@rollup/plugin-alias';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import svelte from 'rollup-plugin-svelte';
import { terser } from 'rollup-plugin-terser';
import del from 'rollup-plugin-delete';
import postcss from "rollup-plugin-postcss";
import autoPreprocess from 'svelte-preprocess'

const IS_PROD = !process.env.ROLLUP_WATCH;
import { getComponentInfo, rollupPluginManifestSync } from "../utils/glue.js";

// External and replacement needs to be the full path :(
export default (key, format) => {
    if (format === undefined) {
        format = "umd";
    }

    const componentKey = key;
    const componentInfo = getComponentInfo(componentKey, !IS_PROD);

    return {
        external: ['shared'],
        input: `src/${componentKey}.js`,
        output: {
            globals: {
                'shared': 'shared',
            },
            sourcemap: !IS_PROD,
            format: format, // if I want to use globals, this is the way
            name: componentInfo.componentKey,
            file: componentInfo.outputPath
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
            postcss({
                extract: true,
            }),

            typescript(),

            svelte({
                dev: !IS_PROD,
                customElement: false,
                exclude: /\.wc\.svelte$/,
                preprocess: autoPreprocess({
                    postcss: true
                }),
                //emitCss: false,
                css: css => {
                    // TODO how to have this cache friendly?
                    css.write(componentInfo.outputPathCSS);
                }
            }),

            // TODO Css is not coming thru when customelement includes non-custom element
            // Possible tip https://github.com/sveltejs/svelte/issues/4274.
            svelte({
                dev: !IS_PROD,
                customElement: true,
                include: /\.wc\.svelte$/,
                preprocess: autoPreprocess({
                    postcss: true
                }),
                css: true, // I Wonder if I actually need this.
            }),

            resolve({
                browser: true,
                dedupe: importee => importee === 'svelte' || importee.startsWith('svelte/')
            }),
            commonjs(),

            /**
             * Minifies JavaScript bundle in production
             */
            IS_PROD && terser(),

            /**
             * Sync the new filename to hugo, for instant feedback
             */
            rollupPluginManifestSync(componentInfo)
        ]
    }
}
