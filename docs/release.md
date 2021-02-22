# Release

```sh
make generate-openapi-js
make generate-openapi-go
make build-site-assets
make build-image-base
make build-image
```


```sh
cat  /srv/learnalist/db/*.sql | sqlite3 /srv/learnalist/server.db
```

```sh
kubectl exec -it $(kubectl get pods -l "app=learnalist" -o jsonpath="{.items[0].metadata.name}") -c learnalist -- \
/app/bin/learnalist-cli --config=/etc/learnalist/config.yaml tools rebuild-static-site
```
