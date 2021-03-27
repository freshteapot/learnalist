import fs from 'fs-extra';
import config from "./src/rollup/setup.js";

const tools = [
    {
        input: "toolbox-language-pad-v1",
        src: "./toolbox/language-pad/v1",
        content: "language-pad-v1",
    },
    {
        input: "toolbox-language-pad-v2",
        src: "./toolbox/language-pad/v2",
        content: "language-pad-v2",
    },
    {
        input: "toolbox-read-write-repeat-v1",
        src: "./toolbox/read-write-repeat/v1",
        content: "read-write-repeat-v1",
    },
    {
        input: "toolbox-plank-v1",
        src: "./plank/index",
        content: "plank-v1",
    }
];

async function hugoTemplate(tool) {
    const template = `
---
title: "Language pad"
type: "toolbox"
layout: "single"
js_include: ["main", "${tool.input}"]
css_include: ["main", "${tool.input}"]
---
`

    const src = `../hugo/content/toolbox/${tool.content}.md`;
    fs.writeFile(src, template.trim());
}

async function writeEntry(tool) {
    const template = `
import Experience from "${tool.src}.svelte";

// Actual app to handle the interactions
let app;
const el = document.querySelector("#main-panel")
if (el) {
    app = new Experience({
        target: el,
    });
}

export default app;
`
    const src = `src/${tool.input}.js`;
    fs.writeFile(src, template);
}

tools.forEach(async tool => {
    await writeEntry(tool);
    await hugoTemplate(tool);
});

const apps = tools.map(tool => {
    return config(tool.input);
});

export default apps;
