```go
inputA, _ := json.Marshal(&aList{Content: "I am a list", UUID: "fake-list-123"})
if err := sc.Publish("publish.alist", inputA); err != nil {
	log.Fatal(err)
}
```

```go
topic := "publish.alist"
if _, err := sc.Subscribe(topic, processAlistFromBytes, stan.DurableName("my-durable")); err != nil {
	log.Fatalf("Failed to start subscription on '%s': %v", topic, err)
}
```

```go
func processAlistsByUserFromBytes(stanMsg *stan.Msg) {
	var lists aListsByUser
	json.Unmarshal(stanMsg.Data, &lists)
	log.Printf("Content: %s, UUID: %s", lists.Content, lists.UserUUID)
}
```
