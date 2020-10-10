# Fixup Log

## Issue 136
### Move planks
- move planks to plank table
- delete plank lists

```sh
go run -tags=json1 main.go --config=../config/dev.config.yaml tools fix-plank-v1 -h
```

### Set lists with no shared object to private

```sh
go run -tags=json1 main.go --config=../config/dev.config.yaml tools fix-acl-owner
```
