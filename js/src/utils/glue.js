const lockfile = require('proper-lockfile');

const fs = require('fs-extra')
const del = require('del');

const pathToManifestFile = "../hugo/data/manifest_js.json";
const pathToManifestFileCSS = "../hugo/data/manifest_css.json";
const pathToStaticDirectory = "../hugo/static";
const pathToPublicDirectory = "../hugo/public";

const getComponentInfo = (componentKey, dev) => {
    let chunkhash = Date.now();
    if (dev) {
        chunkhash = "dev";
    }

    const filename = `${componentKey}.${chunkhash}.js`;
    const filenameCSS = `${componentKey}.${chunkhash}.css`;
    const outputPath = `${pathToStaticDirectory}/js/${filename}`;
    const outputPathCSS = `${pathToStaticDirectory}/css/${filenameCSS}`;

    // Should we only delete dev? and then leave it as a manual step to remove production?
    // Or try and include in rollupdelete?
    const rollupDeleteTargets = [
        `${pathToStaticDirectory}/js/${componentKey}.*.js`,
        `${pathToStaticDirectory}/js/${componentKey}.*.js.map`,
        `${pathToStaticDirectory}/css/${componentKey}.*.css`,
        `${pathToStaticDirectory}/css/${componentKey}.*.css.map`,

        `${pathToPublicDirectory}/js/${componentKey}.*.js`,
        `${pathToPublicDirectory}/js/${componentKey}.*.js.map`,
        `${pathToPublicDirectory}/css/${componentKey}.*.css`,
        `${pathToPublicDirectory}/css/${componentKey}.*.css.map`,
    ];

    return {
        componentKey: componentKey,
        chunkhash: chunkhash,
        filename: filename,
        filenameCSS: filenameCSS,
        outputPath: outputPath,
        outputPathCSS: outputPathCSS,
        rollupDeleteTargets,
    }
}

const write = async (manifestFile, key, value) => {
    console.log(manifestFile)
    const retryOptions = {
        retries: {
            retries: 5,
            factor: 3,
            minTimeout: 1 * 1000,
            maxTimeout: 60 * 1000,
            randomize: true,
        }
    };
    let newManifest = {
        [key]: value
    };

    let file;
    let cleanup;
    try {
        file = '/var/tmp/file.txt';
        await fs.ensureFile(file); // fs-extra creates file if needed
    } catch (err) {
        console.log(err);
        return;
    }

    cleanup = await lockfile.lock(file, retryOptions);

    try {
        await fs.ensureFile(manifestFile)
    } catch (err) {
        // skip error
        console.log(err);
        cleanup();
        return
    }


    try {
        const manifest = await fs.readJson(manifestFile)
        console.log('success!')
        newManifest = { ...manifest, ...newManifest }
    } catch (err) {
        // skip error
        console.log(err);
        cleanup();
        return
    }

    try {
        await fs.writeJson(manifestFile, newManifest, { spaces: ' ' })
        console.log('success!')
    } catch (err) {
        console.log('failed to update manifest!')
        console.log(err);
        // skip error
        cleanup();
        return
    }

    cleanup();
}

const syncManifest = async (componentInfo) => {
    try {
        await write(pathToManifestFile, componentInfo.componentKey, `/js/${componentInfo.filename}`);
        await write(pathToManifestFileCSS, componentInfo.componentKey, `/css/${componentInfo.filenameCSS}`);
    } catch (e) {
        // Deal with the fact the chain failed
        console.log(e)
    }
}

const rollupPluginManifestSync = (componentInfo) => {
    return {
        name: 'sync', // this name will show up in warnings and errors
        generateBundle: () => {
            (async () => {
                try {
                    await syncManifest(componentInfo);
                } catch (e) {
                    // Deal with the fact the chain failed
                    console.log(e)
                }
            })();
        }
    }
}

const syncManifestCSSBase = async () => {
    const chunkhash = Date.now();
    const componentKey = "base";
    const filenameCSS = `${componentKey}.${chunkhash}.css`;
    const outputPathCSS = `${pathToStaticDirectory}/css/${filenameCSS}`;

    const path = '../hugo/public/css/base.min.css';

    try {
        write(pathToManifestFileCSS, componentKey, `/css/${filenameCSS}`);

        const find = `${pathToStaticDirectory}/css/base.*.css`;
        const deletedPaths = await del([find], { dryRun: false, verbose: true, force: true });
        console.log('Files and directories that would be deleted:\n', deletedPaths.join('\n'));
        fs.copySync(path, outputPathCSS);
    } catch (e) {
        // Deal with the fact the chain failed
        console.log(e)
    }
}

export {
    getComponentInfo,
    rollupPluginManifestSync,
    syncManifestCSSBase,
}
