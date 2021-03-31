
CWD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_PWD="$CWD/.."

cd "$ROOT_PWD"
mkdir -p /tmp/learnalist/browser-extension
cp -r js/browser-extension/menu-add-spaced-repetition /tmp/learnalist/browser-extension
cd /tmp/learnalist/browser-extension/menu-add-spaced-repetition
zip -q ../menu-add-spaced-repetition.zip ./*

cat  << _EOF_
Next:
    cd /tmp/learnalist/browser-extension/

Goto:
    https://chrome.google.com/u/0/webstore/devconsole/

- Have you rebuilt the site in production? (will avoid alot of 404 for css and js).
- Once happy, remove stale files from the remote server.

ssh $SSH_SERVER 'sudo bash -s' < /tmp/cleanup/commands
_EOF_
