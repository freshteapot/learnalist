#!/bin/bash
CWD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_PWD="$CWD/.."
build_assets() {
    cd $ROOT_PWD
    cd hugo
    rm -rf ./hugo_stats.json
    rm -rf ./public/*

    HUGO_BUILD_WRITESTATS=true \
    HUGO_PARAMS_BUILDCSS=true \
    HUGO_PARAMS_BUILDCSSPRODUCTION=true \
    hugo --environment=lal01

    ls -lah public/css/base*
    # A second run, to use the "hugo_stats", it would appear
    # Without this, the file in public/css is always just the comments.
    HUGO_BUILD_WRITESTATS=true \
    HUGO_PARAMS_BUILDCSS=true \
    HUGO_PARAMS_BUILDCSSPRODUCTION=true \
    hugo --environment=lal01

    ls -lah public/css/base*
    cd $ROOT_PWD
}

build_js() {
    cd $ROOT_PWD
    cd js
    npm run build:js:components
    cd $ROOT_PWD
}

build() {
    # A little workaround to sync the postcss version from hugo
    # store it in static and provide an entry in hugo/data/manifest_css.json.
    # This is used when we run hugo in production without node / npm / postcss
    # to make rendering of the pages lightning fast still.
    cd $ROOT_PWD
    cd js
    node --experimental-modules sync-site-base-css.mjs

    cd $ROOT_PWD
    cd hugo
    rm -rf ./public/*
    hugo --environment=lal01

    find static
    cd $ROOT_PWD
}

copy_css_classes_from_svelte() {
    # Dump a list of css classes into a hidden file so hugo adds the classes for postcss.
    cd $ROOT_PWD
    echo $PWD
    node ./js/extract-used-css.js > ./hugo/layouts/design/from-svelte.html
}

copy_css_classes_from_svelte
build_js
build_assets
build


# Add static file
#git add hugo/data/manifest_css.json
#git add hugo/data/manifest_js.json
