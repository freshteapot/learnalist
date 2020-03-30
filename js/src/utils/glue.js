const fs = require('fs-extra')

// TODO this might break
const pathToManifestFile = "../hugo/data/manifest.json";
const pathToStaticJSDirectory = "../hugo/static/js";

const getComponentInfo = (componentKey) => {
    const chunkhash = Date.now();
    const filename = `${componentKey}.${chunkhash}.js`;
    const outputPath = `${pathToStaticJSDirectory}/${filename}`;
    const rollupDeleteTargets = [
        `${pathToStaticJSDirectory}/${componentKey}.*.js`,
        `${pathToStaticJSDirectory}/${componentKey}.*.js.map`
    ];

    return {
        componentKey: componentKey,
        chunkhash: chunkhash,
        filename: filename,
        outputPath: outputPath,
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
