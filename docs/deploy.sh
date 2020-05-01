echo "First return:

cd /root/work/src/github.com/freshteapot/learnalist-api/server
git pull --rebase origin master
GO111MODULE=on sh build.sh
cp /srv/learnalist/bin/learnalist-cli /srv/learnalist/learnalist-cli.last.working


Then if happy, run:

mv learnalist-cli /srv/learnalist/bin/learnalist-cli
supervisorctl reload learnalist
sleep 2
supervisorctl status learnalist
supervisorctl tail -5000 learnalist
"
