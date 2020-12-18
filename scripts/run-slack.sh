CWD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TOP_LEVEL="${CWD}/../"
cd $TOP_LEVEL

export EVENTS_VIA="nats"
export EVENTS_STAN_CLUSTER_ID="test-cluster"
export EVENTS_STAN_CLIENT_ID="lal-slack-events"
export EVENTS_NATS_SERVER="127.0.0.1"
export EVENTS_SLACK_WEBHOOK="${EVENTS_SLACK_WEBHOOK:-XXX}"

cd server && \
go run main.go --config=../config/dev.config.yaml tools slack-events
