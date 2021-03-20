#!/bin/bash
CWD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_PWD="$CWD/.."
build_assets() {
    cd $ROOT_PWD
    cd hugo
    echo "Building static site assets"
    rm -rf ./hugo_stats.json
    rm -rf ./public/*

    HUGO_BUILD_WRITESTATS=true \
    HUGO_PARAMS_BUILDCSS=true \
    HUGO_PARAMS_BUILDCSSPRODUCTION=true \
    hugo --quiet --environment=lal01

    #ls -lah public/css/base*
    # A second run, to use the "hugo_stats", it would appear
    # Without this, the file in public/css is always just the comments.
    HUGO_BUILD_WRITESTATS=true \
    HUGO_PARAMS_BUILDCSS=true \
    HUGO_PARAMS_BUILDCSSPRODUCTION=true \
    hugo --quiet --environment=lal01

    #ls -lah public/css/base*
}

build_js() {
    cd $ROOT_PWD
    cd js
    echo "Building js"
    npm run build:js:components
}

build() {
    # A little workaround to sync the postcss version from hugo
    # store it in static and provide an entry in hugo/data/manifest_css.json.
    # This is used when we run hugo in production without node / npm / postcss
    # to make rendering of the pages lightning fast still.
    cd $ROOT_PWD
    cd js
    echo "Syncing css"
    node sync-site-base-css.mjs

    cd $ROOT_PWD
    cd hugo
    rm -rf ./public/*
    hugo --quiet --environment=lal01

    find static
}

copy_css_classes_from_svelte() {
    # Dump a list of css classes into a hidden file so hugo adds the classes for postcss.
    cd $ROOT_PWD
    node ./js/extract-used-css.js > ./hugo/layouts/design/from-svelte.html
}

copy_samples() {
    # We copy samples to make sure all the css is picked up by postcss for each content type
    cd $ROOT_PWD
    cd hugo
    for FOLDER in $(ls samples)
    do
        echo "Copying samples for ${FOLDER}"
        CMD="cp samples/${FOLDER}/*.md content/${FOLDER}/"
        $CMD
        CMD="cp samples/${FOLDER}/*.json data/${FOLDER}/"
        $CMD
    done
}

remove_samples() {
    # We remove samples as they are not needed on the site
    cd $ROOT_PWD
    cd hugo
    #rm content/challenge/*.md
    #rm data/challenge/*.json
    for FOLDER in $(ls samples)
    do
        echo "Removing samples for ${FOLDER}"
        CMD="rm content/${FOLDER}/*.md"
        $CMD
        CMD="rm data/${FOLDER}/*.json"
        $CMD
    done

}

copy_samples
copy_css_classes_from_svelte
build_js
build_assets
build
remove_samples
