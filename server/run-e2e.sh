#!/usr/bin/env bash
rm "/tmp/learnalist/e2e-all.log"
rm "/tmp/learnalist/e2e.log"

cd ./e2e

go clean -testcache && \
go test --tags="json1" \
-ginkgo.v \
-ginkgo.progress \
-ginkgo.focus="Smoke list access|Static Site Simple flow" \
-test.v .

cat "/tmp/learnalist/e2e.log" >> "/tmp/learnalist/e2e-all.log"

# We do this last, as its quite intense
# I think the issue is the amount of events hitting static-site slows down the build
go clean -testcache && \
go test --tags="json1" \
-ginkgo.v \
-ginkgo.progress \
-ginkgo.skip="Smoke list access|Static Site Simple flow" \
-test.v .
cat "/tmp/learnalist/e2e.log" >> "/tmp/learnalist/e2e-all.log"
