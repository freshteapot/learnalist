# Manually Install on the server

Maybe get a ubuntu local copy so I dont need to do it from the server.

## Update the learnalist server
```
sudo su -
cd /root/work/src/github.com/freshteapot/learnalist-api/
git pull --rebase origin master
cd api/
GO111MODULE=on sh build.sh
```
Make a backup of the one running
```
cp /root/work/bin/api api.last.working
```

Move it to where supervisor will find it.
```
mv apiserver /root/work/bin/api
```
When ready, reload
```
supervisorctl reload learnalist-api
```


## Change golang
```
cd /root
rm go1.6.linux-amd64.tar.gz
rm go*
mkdir tmp
cd tmp
curl -O 'https://dl.google.com/go/go1.12.4.linux-amd64.tar.gz'
tar -zxf go1.12.4.linux-amd64.tar.gz
rm -rf /usr/local/go
mv go /usr/local
cd ..
rm -rf tmp
```
