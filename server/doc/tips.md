# Tips from the coal Face

## Turn []byte into a string for easy reading
```go
  buf := new(bytes.Buffer)
  buf.ReadFrom(body)
  s := buf.String()
  fmt.Println(s)
```

## Working with structs and json

Get the Data object and add a single row to the v2 type data.
```go
aListV2Data := aList.Data.(alist.TypeV2)

item := &alist.TypeV2Item{From: "Hi", To: "Hello"}
aListV2Data = append(aListV2Data, *item)
aList.Data = aListV2Data
```

# Working with casbin
p = policy

Get all policies, loop over them and get the policy detail.
```go
allNamedObjects := acl.enforcer.GetAllNamedObjects("p")
for _, policy := range allNamedObjects {
  fmt.Println(policy)
  filteredPolicy = acl.enforcer.GetFilteredPolicy(1, policy)
  fmt.Println(filteredPolicy)
}
```

Filter policy on position 2 to get action

```
items := acl.enforcer.GetFilteredPolicy(2, "read")
```

Filter policy on position 1 to get obj

```
items := acl.enforcer.GetFilteredPolicy(1)
```


# To run go run with the vendor folder

```sh
go mod download
go mod vendor
go run -mod=vendor *
```

# To run smoke test

```sh
cd e2e
go test .
```

# Make smoke test binary
```sh
cd e2e
go test -o smoke-test -c
```
