
CWD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_PWD="$CWD/.."

cd "$ROOT_PWD"
rm -rf /tmp/learnalist/browser-extension
mkdir -p /tmp/learnalist/browser-extension
cp -r js/browser-extension/menu-add-spaced-repetition /tmp/learnalist/browser-extension
cd /tmp/learnalist/browser-extension/menu-add-spaced-repetition
zip -rq ../menu-add-spaced-repetition.zip ./*
cd /tmp/learnalist/browser-extension/
mkdir temp
cp menu-add-spaced-repetition.zip temp
cd temp
unzip menu-add-spaced-repetition.zip


cat  << _EOF_
Next:
    cd /tmp/learnalist/browser-extension/

Goto:
    https://chrome.google.com/u/0/webstore/devconsole/
_EOF_
