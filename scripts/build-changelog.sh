CWD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_PWD="$CWD/.."

build_changelog() {
    cd ~/git/git-log-json
    go run main.go $ROOT_PWD | jq > "$ROOT_PWD/hugo/data/changelog.json"
}

build_changelog
