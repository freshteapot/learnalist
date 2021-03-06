import lockfile from 'proper-lockfile';
import fs from 'fs-extra';
import del from 'del';

const basePath = "../hugo"
const localBasePath = "dist"
const pathToManifestFile = `${basePath}/data/manifest_js.json`;
const pathToManifestFileCSS = `${basePath}/data/manifest_css.json`;
const pathToStaticDirectory = `${basePath}/static`;
const pathToPublicDirectory = `${basePath}/public`;

const getComponentInfo = (componentKey, production) => {
    let chunkhash = production ? "." + Date.now() : "";

    const filenameJS = `${componentKey}${chunkhash}.js`;
    const filenameCSS = `${componentKey}${chunkhash}.css`;

    // Should we only delete dev? and then leave it as a manual step to remove production?
    // Or try and include in rollupdelete?
    const rollupDeleteTargets = [
        // Delete local
        `${localBasePath}/${componentKey}.*`,

        // Delete staticsite: hugo static
        `${pathToStaticDirectory}/js/${componentKey}.*`,
        `${pathToStaticDirectory}/css/${componentKey}.*`,

        // Development only
        // Delete staticsite: hugo public
        `${pathToPublicDirectory}/js/${componentKey}.*`,
        `${pathToPublicDirectory}/css/${componentKey}.*`,
    ];

    // Horrible, for now
    const rollupCopyTargets = [
        { src: `dist/${componentKey}.js`, dest: `${pathToStaticDirectory}/js/`, rename: `${filenameJS}` },
        { src: `dist/${componentKey}.js.map`, dest: `${pathToStaticDirectory}/js/`, rename: `${filenameJS}.map` },
        { src: `dist/${componentKey}.css`, dest: `${pathToStaticDirectory}/css/`, rename: `${filenameCSS}` },
    ];

    return {
        componentKey,
        chunkhash,
        filenameJS,
        filenameCSS,
        localBasePath,

        rollupDeleteTargets,
        rollupCopyTargets,
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
        newManifest = { ...manifest, ...newManifest }
    } catch (err) {
        // skip error
        console.log(err);
        cleanup();
        return
    }

    try {
        await fs.writeJson(manifestFile, newManifest, { spaces: ' ' })
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
        await write(pathToManifestFile, componentInfo.componentKey, `/js/${componentInfo.filenameJS}`);
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
