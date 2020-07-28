const fs = require('fs');
const glob = require("glob")
const { preprocess } = require('svelte/compiler');
const DIR = __dirname;

async function getCssClasses(filename) {
    let cssClasses = [];
    const source = fs.readFileSync(filename, 'utf-8');
    const { code, dependencies } = preprocess(source, {
        markup: async ({ content, filename }) => {
            cssClasses = require("purgecss-from-svelte").extract(content)
        },
    });
    return cssClasses;
}

function onlyUnique(value, index, self) {
    return self.indexOf(value) === index;
}

// TODO would be nice to not include tags
// TODO would be nice to only extract out tachyons
glob(`${DIR}/src/**/*.svelte`, async function (er, files) {
    const cssClasses = [];

    await Promise.all(files.map(async (filename) => {
        try {
            const c = await getCssClasses(filename);
            cssClasses.push(...c);
        } catch (error) {
            console.log('error' + error);
        }
    }))

    var unique = cssClasses.filter(onlyUnique);
    // Write this to a tag with class, and let hugo inject all the css
    console.log(`<div class="${unique.join(" ")}"/>`);
})

