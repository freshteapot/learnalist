# Monitoring



## Read the challenges stream
### Setup a tunnel
```
ssh $SSH_SERVER -L 4222:127.0.0.1:4222 -N &
ssh $SSH_SERVER sudo kubectl port-forward deployment/stan01 4222:4222 &
```

### Consume challenges stream
```
cd server
TOPIC=challenges \
EVENTS_STAN_CLIENT_ID=nats-reader \
EVENTS_STAN_CLUSTER_ID=stan \
EVENTS_NATS_SERVER=127.0.0.1 \
go run main.go --config=../config/dev.config.yaml \
tools natsutils read
```


# How many users with tokens + name

```
SELECT
	uuid,
IFNULL(json_extract(body, '$.display_name'), uuid) AS display_name
FROM
	user_info
WHERE
	uuid IN(SELECT DISTINCT user_uuid FROM mobile_device);
```
