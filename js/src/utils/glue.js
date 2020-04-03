const fs = require('fs-extra')

// TODO this might break
const pathToManifestFile = "../hugo/data/manifest.json";
const pathToManifestFileCSS = "../hugo/data/manifest_css.json";
const pathToStaticJSDirectory = "../hugo/static";

const getComponentInfo = (componentKey) => {
    const chunkhash = Date.now();
    const filename = `${componentKey}.${chunkhash}.js`;
    const filenameCSS = `${componentKey}.${chunkhash}.css`;
    const outputPath = `${pathToStaticJSDirectory}/js/${filename}`;
    const outputPathCSS = `${pathToStaticJSDirectory}/css/${filenameCSS}`;
    const rollupDeleteTargets = [
        `${pathToStaticJSDirectory}/js/${componentKey}.*.js`,
        `${pathToStaticJSDirectory}/js/${componentKey}.*.js.map`,
        `${pathToStaticJSDirectory}/css/${componentKey}.*.css`,
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
    let newManifest = {
        [key]: value
    };

    try {
        await fs.ensureFile(manifestFile)
    } catch (err) {
        // skip error
    }


    try {
        manifest = await fs.readJson(manifestFile)
        console.log('success!')
        newManifest = { ...manifest, ...newManifest }
    } catch (err) {
        // skip error
    }

    try {
        await fs.writeJson(manifestFile, newManifest)
        console.log('success!')
    } catch (err) {
        console.log('failed to update manifest!')
        // skip erro
    }
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


module.exports = {
    getComponentInfo,
    rollupPluginManifestSync,
}
