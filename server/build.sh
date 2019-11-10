pathToVersion="github.com/freshteapot/learnalist-api/server/api/version"
gitHash=$(git rev-parse HEAD)
buildVersion="v0.0.1"
# This is in UTC
gitHashDate=$(TZ=UTC git show --quiet --date='format-local:%Y%m%dT%H%M%SZ' --format="%cd" ${gitHash})
cmd=$(cat <<_EOF_
go build -ldflags "-s -w " -ldflags "
-X ${pathToVersion}.GitHash=${gitHash}
-X ${pathToVersion}.GitDate=${gitHashDate}
-X ${pathToVersion}.Version=${buildVersion}
" -o learnalist-cli main.go
_EOF_
)
echo "Will run the command:"
echo $cmd
echo ""
eval $cmd
