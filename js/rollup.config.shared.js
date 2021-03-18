import typescript from '@rollup/plugin-typescript';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import svelte from 'rollup-plugin-svelte';
import { terser } from 'rollup-plugin-terser';
import del from 'rollup-plugin-delete';
import sveltePreprocess from 'svelte-preprocess';
import copy from 'rollup-plugin-copy'

const production = !process.env.ROLLUP_WATCH;

import { getComponentInfo, rollupPluginManifestSync } from "./src/utils/glue.mjs";
const componentKey = "shared";
const componentInfo = getComponentInfo(componentKey, !production);

export default {
    input: 'src/shared.js',
    output: {
        name: componentInfo.componentKey,
        sourcemap: !production,
        format: 'iife',
        dir: componentInfo.localBasePath,
    },
    plugins: [
        del({ targets: componentInfo.rollupDeleteTargets, verbose: true, force: true }),
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
                customElement: true
            },
        }),

        copy({
            targets: componentInfo.rollupCopyTargets,
            verbose: true,
            force: true,
            hook: 'writeBundle'
        }),

        resolve(),
        commonjs(),

        /**
         * Minifies JavaScript bundle in production
         */
        production && terser(),

        /**
         * Sync the new filename to hugo, for instant feedback
         */
        rollupPluginManifestSync(componentInfo)
    ]
};
