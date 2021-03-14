#!/usr/bin/env bash
rm "/tmp/learnalist/e2e-all.log"
rm "/tmp/learnalist/e2e.log"

cd ./e2e
go clean -testcache && \
go test --tags="json1" \
-ginkgo.v \
-ginkgo.progress \
-test.v .
#-ginkgo.skip="Smoke list access|Static Site Simple flow" \

cat "/tmp/learnalist/e2e.log" >> "/tmp/learnalist/e2e-all.log"
# We do this after, as the above is quite intense
sleep 2
exit

go clean -testcache && \
go test --tags="json1" \
-ginkgo.v \
-ginkgo.progress \
-ginkgo.focus="Smoke list access|Static Site Simple flow" \
-test.v .

cat "/tmp/learnalist/e2e.log" >> "/tmp/learnalist/e2e-all.log"
