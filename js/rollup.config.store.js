import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import svelte from 'rollup-plugin-svelte';
import { terser } from 'rollup-plugin-terser';
import del from 'rollup-plugin-delete';

const IS_PROD = !process.env.ROLLUP_WATCH;

const { getComponentInfo, rollupPluginManifestSync } = require("./src/utils/glue.js");
const componentKey = "shared";
const componentInfo = getComponentInfo(componentKey);

export default {
    input: 'src/store.js',
    output: {
        name: "superstore",
        sourcemap: !IS_PROD,
        format: 'iife',
        file: componentInfo.outputPath
    },
    plugins: [
        del({ targets: componentInfo.rollupDeleteTargets, verbose: true, force: true }),
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
