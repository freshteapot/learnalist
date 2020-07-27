#!/bin/sh
# Set LAL_BIND to bing hugo and server to none localhost
#
# LAL_BIND="192.168.0.10" make develop
#
# trap ctrl-c and call ctrl_c()
trap ctrl_c INT

function ctrl_c() {
	echo "** Trapped CTRL-C"
	sleep 2
	kill -9 $(lsof -ti tcp:1313) 2>/dev/null
	kill -9 $(lsof -ti tcp:1234) 2>/dev/null
}

kill -9 $(lsof -ti tcp:1313) 2>/dev/null
kill -9 $(lsof -ti tcp:1234) 2>/dev/null

CWD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

HUGO_DIR="${CWD}/../hugo"

#LAL_BIND="192.168.0.10"
SERVER_CONFIG="../config.dev_external.yaml"
#HUGO_ARGS_BASEURL="http://192.168.0.10:1313" HUGO_ARGS_BIND="192.168.0.10" make develop2

DEFAULT_SERVER_CONFIG="${CWD}/../config/dev.config.yaml"
_BIND="${LAL_BIND:-localhost}"
_BASEURL="http://${_BIND}:1313"
_APISERVER="http://${_BIND}:1234"
SERVER_CONFIG="${CWD}/../config/dev_external.config.yaml"
HUGO_CONFIG_DIR="${HUGO_DIR}/config/dev_external"
HUGO_CONFIG="${HUGO_CONFIG_DIR}/config.yaml"

# Notice
echo "Running hugo on ${_APISERVER} with config from ${HUGO_CONFIG}."
echo "Running server with config from ${SERVER_CONFIG}."

# Update config server
rm -f $SERVER_CONFIG
cp "${CWD}/../config/dev.config.yaml" $SERVER_CONFIG
yq w -i $SERVER_CONFIG hugo.environment "dev_external"
yq w -i $SERVER_CONFIG server.cookie.domain "${_BIND}"

# Update config hugo
mkdir -p $HUGO_CONFIG_DIR
rm -f $HUGO_CONFIG
touch $HUGO_CONFIG
yq w -i $HUGO_CONFIG baseURL  "${_BASEURL}"
yq w -i $HUGO_CONFIG params.ApiServer "${_APISERVER}"
#yq w -i "$HUGO_DIR/config/dev_external/config.yaml" params.apiServer2 "${_APISERVER}"

mkdir -p "$HUGO_DIR/public"
rm -rf "$HUGO_DIR/public/"*
ls -lah "$HUGO_DIR/public"
cd server && \
go run --tags="json1" main.go --config=$SERVER_CONFIG server &


cd $HUGO_DIR && \
hugo server \
-w \
-e dev_external \
--disableFastRender  \
--forceSyncStatic \
--renderToDisk \
--verbose \
--verboseLog \
--debug \
-b $_BASEURL --bind $_BIND \
&
sleep 1
