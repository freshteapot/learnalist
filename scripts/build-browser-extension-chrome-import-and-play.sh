
CWD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_PWD="$CWD/.."

cd "$ROOT_PWD"
rm -rf /tmp/learnalist/browser-extension
mkdir -p /tmp/learnalist/browser-extension

cd "$ROOT_PWD/js"
npm run build:js:browser-extension:import-play

cd "$ROOT_PWD"
cp -r js/browser-extension/import-play /tmp/learnalist/browser-extension

cd /tmp/learnalist/browser-extension/import-play
zip -r ../import-play.zip ./*
cd /tmp/learnalist/browser-extension/
mkdir temp
cp import-play.zip temp
cd temp
unzip import-play.zip


cat  << _EOF_
Next:
    cd /tmp/learnalist/browser-extension/

Goto:
    https://chrome.google.com/u/0/webstore/devconsole/
_EOF_
