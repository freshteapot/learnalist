echo "First return:

cd /root/work/src/github.com/freshteapot/learnalist-api/server
git pull --rebase origin master
GO111MODULE=on sh build.sh
cp /root/work/bin/api api.last.working


Then if happy, run:

mv apiserver /root/work/bin/api
supervisorctl reload learnalist-api
sleep 2
supervisorctl status learnalist-api
supervisorctl tail -5000 learnalist-api
"
