echo "# Output from the tests"
go test `go list ./... | grep -vE 'integrations|e2e'` -covermode=count -coverprofile=profile.cov
echo ""
TOTAL_COVERAGE=$(go tool cover -func=profile.cov | tail -1 | awk '{gsub("%","",$3)} {print $3}')
OUTPUT=$(cat <<-END
# Total Coverage
${TOTAL_COVERAGE}%

# Make a new badge
gopherbadger -manualcov="${TOTAL_COVERAGE}"

# View output in the browser
go tool cover -html=profile.cov

END
)
echo "${OUTPUT}"
