# Install server for development

## Setup
- setup folder structure and remove any data that already exists
- setup empty database

```sh
make clear-site
make rebuild-db
```

##  Run the server
-
```sh
make run-api-server
```

## Run the server and run hugo from within
```sh
HUGO_EXTERNAL=false make run-api-server
```

## Run hugo, server, js
- this will use hugo externally
- hugo on port 1313
- server on port 1234
```sh
make develop
```
