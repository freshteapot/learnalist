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
	kill -9 $(lsof -ti tcp:1313) >/dev/null 2>&1
	kill -9 $(lsof -ti tcp:1234) >/dev/null 2>&1
}

function check_installed() {
	npm -v >/dev/null 2>&1
	installed=$?
	if [[ $installed != 0 ]]; then
		echo "npm needs to be installed, make sure node is as well"
		exit 1
	fi

	hugo version >/dev/null 2>&1
	installed=$?
	if [[ $installed != 0 ]]; then
		echo "hugo needs to be installed"
		exit 1
	fi

	yq --version >/dev/null 2>&1
	installed=$?
	if [[ $installed != 0 ]]; then
		echo "yq needs to be installed"
		exit 1
	fi

	go version >/dev/null 2>&1
	installed=$?
	if [[ $installed != 0 ]]; then
		echo "go needs to be installed"
		exit 1
	fi
}

# Make sure the ports are killed before we start and hopefully catch them on close
kill -9 $(lsof -ti tcp:1313) >/dev/null 2>&1
kill -9 $(lsof -ti tcp:1234) >/dev/null 2>&1
# Poor attempt at confirming we have the commands installed
check_installed

# Config setup
_BIND="${LAL_BIND:-$(ipconfig getifaddr en0)}"
_BASEURL="http://${_BIND}:1313"
_APISERVER="http://${_BIND}:1234"

CWD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TOP_LEVEL="${CWD}/../"
HUGO_DIR="${CWD}/../hugo"
DEFAULT_SERVER_CONFIG="${CWD}/../config/dev.config.yaml"
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
yq w -i $HUGO_CONFIG baseURL "${_BASEURL}"
yq w -i $HUGO_CONFIG params.ApiServer "${_APISERVER}"

# Clean up hugo output
mkdir -p "$HUGO_DIR/public"
rm -rf "$HUGO_DIR/public/"*
mkdir -p $HUGO_DIR/{public/alist,public/alistsbyuser}
ls -lah "$HUGO_DIR/public"
# Start the server

cd server && \
STATIC_SITE_EXTERNAL=false \
go run --tags="json1" main.go --config=$SERVER_CONFIG server &

# Start static site engine
cd $HUGO_DIR && \
hugo server \
-w \
-e dev_external \
--disableFastRender \
--forceSyncStatic \
--renderToDisk \
--verbose \
--verboseLog \
--debug \
-b $_BASEURL --bind $_BIND \
&
sleep 1


cd $TOP_LEVEL
cd js
npm run dev:js:components
