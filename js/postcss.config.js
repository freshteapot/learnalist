let production = !process.env.ROLLUP_WATCH;
const purgeCss = require("@fullhuman/postcss-purgecss")({
    content: ["./src/**/*.svelte"],
    defaultExtractor: content => [
        ...(content.match(/[^<>"'`\s]*[^<>"'`\s:]/g) || []),
        ...(content.match(/(?<=class:)[^=>\/\s]*/g) || []),
    ],
})

module.exports = {
    plugins: [
        require("postcss-import")(),
        production && purgeCss
    ]
};
