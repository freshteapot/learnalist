# Manually Install on the server

Maybe get a ubuntu local copy so I dont need to do it from the server.

## Update the learnalist server
```sh
sudo su -
cd /root/work/src/github.com/freshteapot/learnalist-api/
git pull --rebase origin master
cd server/
GO111MODULE=on sh build.sh
```

Create the files for hugo
```
mkdir -p /var/www/learnalist-api/site-cache
cp -r $(pwd)/alists/hugo /var/www/learnalist-api/hugo
mkdir -p /var/www/learnalist-api/hugo/{public-alist,content/alists,data/lists}
chown -R www-data:www-data /var/www/learnalist-api/
```

```
chown -R www-data:www-data /var/www/learnalist-api/
```

Make a backup of the one running
```sh
cp /root/work/bin/server server.last.working
```

Move it to where supervisor will find it.
```sh
mv apiserver /root/work/bin/server
```
When ready, reload
```sh
supervisorctl reload learnalist-api
```

Check the logs
```
supervisorctl tail -f  learnalist-api
```


## Change golang
```sh
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

## Update the database with all changes.
```sh
ls db/*.sql | sort | xargs cat | sqlite3 server.db
```

## Update the database with a single file change.
```sh
cat  db/201905052144-labels.sql | sqlite3 test.db
```
