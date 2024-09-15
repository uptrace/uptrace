module github.com/vmihailenco/taskq/example/redisexample

go 1.22

toolchain go1.23.0

replace github.com/vmihailenco/taskq/v4 => ../..

replace github.com/vmihailenco/taskq/redisq/v4 => ../../redisq

replace github.com/vmihailenco/taskq/memqueue/v4 => ../../memqueue

replace github.com/vmihailenco/taskq/taskqtest/v4 => ../../taskqtest

require (
	github.com/redis/go-redis/v9 v9.0.3
	github.com/vmihailenco/taskq/redisq/v4 v4.0.0-beta.4
	github.com/vmihailenco/taskq/v4 v4.0.0-beta.4
)

require (
	github.com/bsm/redislock v0.9.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-redis/redis_rate/v10 v10.0.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/klauspost/compress v1.16.3 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
)
