
let production = !process.env.ROLLUP_WATCH;
const purgeCss = require("@fullhuman/postcss-purgecss")({
    content: ["./src/**/*.svelte"],
    //defaultExtractor: (content) => content.match(/[A-Za-z0-9-_:/]+/g) || [],
    defaultExtractor: content => [
        ...(content.match(/[^<>"'`\s]*[^<>"'`\s:]/g) || []),
        ...(content.match(/(?<=class:)[^=>\/\s]*/g) || []),
    ],
})

module.exports = {
    plugins: [
        require("postcss-import")(),
        //production && purgeCss
        purgeCss
    ]
};
