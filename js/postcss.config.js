let production = !process.env.ROLLUP_WATCH;
const purgeCss = require("@fullhuman/postcss-purgecss")({
    content: ["./src/**/*.svelte"],
    css: ["**/*.css"],
    extractors: [
        {
            extractor: content => require("purgecss-from-svelte").extract(content),
            extensions: ["svelte"]
        }
    ]
})

module.exports = {
    plugins: [
        require("postcss-import")(),
        require("autoprefixer"),
        // Only purge css on production
        production && purgeCss
    ]
};
