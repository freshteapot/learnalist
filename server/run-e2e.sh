#!/usr/bin/env bash
cd ./e2e
go clean -testcache && go test --tags="json1" -ginkgo.v -ginkgo.progress -ginkgo.skip="Smoke list access"  -ginkgo.skip="Smoke user" -test.v .
# We do this after, as the above is quite intense
sleep 2
go clean -testcache && go test --tags="json1" -ginkgo.v -ginkgo.progress -ginkgo.focus="Smoke list access" -test.v .
