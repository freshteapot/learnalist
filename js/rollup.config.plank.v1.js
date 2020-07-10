import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import svelte from 'rollup-plugin-svelte';
import { terser } from 'rollup-plugin-terser';
import del from 'rollup-plugin-delete';
import postcss from "rollup-plugin-postcss";
import autoPreprocess from 'svelte-preprocess'

const IS_PROD = !process.env.ROLLUP_WATCH;
const { getComponentInfo, rollupPluginManifestSync } = require("./src/utils/glue.js");
const componentKey = "plank-v1";
const componentInfo = getComponentInfo(componentKey, !IS_PROD);

// External and replacement needs to be the full path :(
export default {
  input: 'src/plank-v1.js',
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
      customElement: false,
      exclude: /\.wc\.svelte$/,
      preprocess: autoPreprocess({
        postcss: true
      }),
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
      browser: true
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
