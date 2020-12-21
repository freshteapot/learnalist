#!/bin/bash
cd server && \
go run main.go --config=../config/dev.config.yaml tools stub-sql-file $1
