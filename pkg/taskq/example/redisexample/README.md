# taskq example using Redis backend

The example requires Redis Server running on the `:6379`.

First, start the consumer:

```shell
go run consumer/main.go

```

Then, start the producer:

```shell
go run producer/main.go
```
