# Manually Install on the server

## Setup

### Setup log files
```
mkdir -p /var/log/learnalist/
chown -R root:root /var/log/learnalist/
```

### Copy config
```
cp config/prod.config.yaml /srv/learnalist/
```

### Setup supervisor
```
cp config/supervisor.conf.learnalist.conf /etc/supervisor/conf.d/learnalist.conf
```


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
```sh
cd ..
sudo -u www-data -g www-data mkdir -p /srv/learnalist/{bin,site-cache}
sudo -u www-data -g www-data cp -rf ./hugo/static/* /srv/learnalist/site-cache/
sudo -u www-data -g www-data cp -r ./hugo /srv/learnalist
sudo -u www-data -g www-data mkdir -p /srv/learnalist/hugo/{public,content/alist,data/alist,content/alistsbyuser,data/alistsbyuser}
chown -R www-data:www-data /srv/learnalist/
```

Make a backup of the one running

```sh
cp /srv/learnalist/bin/learnalist-cli /srv/learnalist/learnalist-cli.last.working
```

Move it to where supervisor will find it.
```sh
cp -f server/learnalist-cli /srv/learnalist/bin/learnalist-cli
```
When ready, reload
```sh
supervisorctl reload learnalist
```

Check the logs
```
supervisorctl tail -f  learnalist
```

Rebuild the site, if you know javascript, templates changed
```
sudo -uwww-data /srv/learnalist/bin/learnalist-cli tools rebuild-static-site --config=/srv/learnalist/prod.config.yaml
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
