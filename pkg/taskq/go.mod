module github.com/vmihailenco/taskq/v4

go 1.22

replace github.com/vmihailenco/taskq/memqueue/v4 => ./memqueue

require (
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13
	github.com/go-logr/logr v1.2.4
	github.com/go-logr/stdr v1.2.2
	github.com/go-redis/redis_rate/v10 v10.0.1
	github.com/hashicorp/golang-lru v0.5.4
	github.com/klauspost/compress v1.16.3
	github.com/redis/go-redis/v9 v9.0.3
	github.com/vmihailenco/msgpack/v5 v5.3.5
	github.com/vmihailenco/taskq/memqueue/v4 v4.0.0-beta.4
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
)
