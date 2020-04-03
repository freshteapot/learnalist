import alias from '@rollup/plugin-alias';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import svelte from 'rollup-plugin-svelte';
import { terser } from 'rollup-plugin-terser';
import del from 'rollup-plugin-delete';
import postcss from "rollup-plugin-postcss";
import autoPreprocess from 'svelte-preprocess'

const IS_PROD = !process.env.ROLLUP_WATCH;
const { getComponentInfo, rollupPluginManifestSync } = require("./src/utils/glue.js");
const componentKey = "editor";
const componentInfo = getComponentInfo(componentKey);

// External and replacement needs to be the full path :(
export default {

  input: 'src/editor.js',
  output: {

    sourcemap: !IS_PROD,
    format: 'esm',
    name: componentInfo.componentKey,
    file: componentInfo.outputPath
  },
  plugins: [
    del({ targets: componentInfo.rollupDeleteTargets, verbose: true, force: true }),
    postcss({
      extract: true,
    }),
    svelte({
      dev: !IS_PROD,
      preprocess: autoPreprocess({
        postcss: true
      })
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
};
