#!/bin/bash
# TODO pimp the shit out of this via hugo
# Reference
# https://apihandyman.io/api-toolbox-jq-and-openapi-part-1-using-jq-to-extract-data-from-openapi-files/
ORIGPWD=$PWD

OPENAPI_SPEC="${ORIGPWD}/learnalist.yaml"
OUTPUT="${ORIGPWD}/docs/api.auto.md"

cd scripts
rm -rf $OUTPUT
touch $OUTPUT
echo "# Api Poor mans auto generation" >> $OUTPUT
yq r $OPENAPI_SPEC -j | jq -r -f list-operations.jq >> $OUTPUT
cd $ORIGPWD
