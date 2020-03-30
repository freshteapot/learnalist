const purgecss = require('@fullhuman/postcss-purgecss')

module.exports = {
    plugins: [
        require("postcss-import")(),
        require("autoprefixer"),
        // Only purge css on production
        purgecss({
            content: ['./**/*.html']
        })
    ]
};
