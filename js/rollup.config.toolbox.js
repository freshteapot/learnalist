import fs from 'fs-extra';
import config from "./src/rollup/setup.js";

const tools = [
    {
        input: "toolbox-language-pad-v1",
        src: "./toolbox/language-pad/v1",
        content: {
            stub: "language-pad-v1",
            title: "Language Pad",
        },
    },
    {
        input: "toolbox-read-write-repeat-v1",
        src: "./toolbox/read-write-repeat/v1",
        content: {
            stub: "read-write-repeat-v1",
            title: "Read Write Repeat"
        },
    },
    {
        input: "toolbox-plank-stats",
        src: "./plank/stats/v1/v1",
        content: {
            stub: "plank-stats-v1",
            title: "Plank stats",
        },
    },
    {
        input: "toolbox-plank-stats-v2",
        src: "./plank/stats/v2/v2",
        content: {
            stub: "plank-stats-v2",
            title: "Have I planked Today",
        },
    }
];

async function hugoTemplate(tool) {
    const template = `
---
title: "${tool.content.title}"
type: "toolbox"
layout: "single"
js_include: ["main", "${tool.input}"]
css_include: ["main", "${tool.input}"]
---
`

    const src = `../hugo/content/toolbox/${tool.content.stub}.md`;
    fs.writeFile(src, template.trim());
}

async function writeEntry(tool) {
    const template = `
// Auto generated from rollup.config.toolbox.js
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
    fs.writeFile(src, template.trim());
}

tools.forEach(async tool => {
    await writeEntry(tool);
    await hugoTemplate(tool);
});

const apps = tools.map(tool => {
    return config(tool.input);
});

export default apps;
