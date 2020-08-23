import typescript from '@rollup/plugin-typescript';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import svelte from 'rollup-plugin-svelte';
import { terser } from 'rollup-plugin-terser';
import del from 'rollup-plugin-delete';

const IS_PROD = !process.env.ROLLUP_WATCH;

import { getComponentInfo, rollupPluginManifestSync } from "./src/utils/glue.mjs";
const componentKey = "shared";
const componentInfo = getComponentInfo(componentKey, !IS_PROD);

export default {
    input: 'src/shared.js',
    output: {
        name: "shared",
        sourcemap: !IS_PROD,
        format: 'iife',
        file: componentInfo.outputPath
    },
    plugins: [
        del({ targets: componentInfo.rollupDeleteTargets, verbose: true, force: true }),
        typescript(),
        svelte({
            dev: !IS_PROD,
            customElement: true,
            css: false
        }),

        resolve(),
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
};
