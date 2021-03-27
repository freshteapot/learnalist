#
# Based on data in the manifests, build a list of files to remove that are out of date
#

if [ -z "$SSH_SERVER" ]; then
    echo "SSH_SERVER is not set"
    exit
fi

rm -rf /tmp/cleanup
mkdir /tmp/cleanup
cd /tmp/cleanup

# Build a list of files to tame
touch static-files.txt
ssh $SSH_SERVER -C 'find /srv/learnalist/hugo/static/css' >> static-files.txt
ssh $SSH_SERVER -C 'find /srv/learnalist/hugo/static/js' >> static-files.txt
ssh $SSH_SERVER -C 'find /srv/learnalist/hugo/public/css' >> static-files.txt
ssh $SSH_SERVER -C 'find /srv/learnalist/hugo/public/js' >> static-files.txt
scp $SSH_SERVER:/srv/learnalist/hugo/data/manifest_css.json .
scp $SSH_SERVER:/srv/learnalist/hugo/data/manifest_js.json .

touch commands
for manifest in "manifest_css.json"  "manifest_js.json"; do
    LINES=$(cat $manifest | jq -rc 'to_entries[]')


    for line in $LINES; do
        KEY=$(echo $line | jq -r '.key')
        current=$(echo $line | jq -r '.value')
        prefix=$(echo $current | awk -F'.' '{print $1}')

        cat static-files.txt| grep $prefix | grep -v $current | awk '{cmd="rm "$0;print(cmd)}' >> commands
    done
done

cat  << _EOF_
- Look in /tmp/cleanup/commands.
- Have you rebuilt the site in production? (will avoid alot of 404 for css and js).
- Once happy, remove stale files from the remote server.

ssh $SSH_SERVER 'sudo bash -s' < /tmp/cleanup/commands
_EOF_
