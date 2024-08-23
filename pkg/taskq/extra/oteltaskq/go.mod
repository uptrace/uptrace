module github.com/vmihailenco/taskq/extra/oteltaskq/v4

go 1.20

replace github.com/vmihailenco/taskq/v4 => ../..

replace github.com/vmihailenco/taskq/memqueue/v4 => ../../memqueue

require (
	github.com/vmihailenco/taskq/v4 v4.0.0-beta.4
	go.opentelemetry.io/otel v1.15.1
	go.opentelemetry.io/otel/trace v1.15.1

)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-redis/redis_rate/v10 v10.0.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/klauspost/compress v1.16.3 // indirect
	github.com/redis/go-redis/v9 v9.0.3 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
)
