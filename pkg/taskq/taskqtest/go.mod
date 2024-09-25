module github.com/vmihailenco/taskq/taskqtest/v4

go 1.22

replace github.com/vmihailenco/taskq/v4 => ./..

replace github.com/vmihailenco/taskq/memqueue/v4 => ../memqueue

require (
	github.com/go-redis/redis_rate/v10 v10.0.1
	github.com/redis/go-redis/v9 v9.0.3
	github.com/stretchr/testify v1.8.2
	github.com/vmihailenco/taskq/v4 v4.0.0-beta.4
)

require (
	github.com/bsm/ginkgo/v2 v2.7.0 // indirect
	github.com/bsm/gomega v1.26.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/ianlancetaylor/demangle v0.0.0-20200824232613-28f6c0f3b639 // indirect
	github.com/klauspost/compress v1.16.3 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/ginkgo/v2 v2.9.2 // indirect
	github.com/onsi/gomega v1.27.6 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/vmihailenco/taskq/memqueue/v4 v4.0.0-beta.4 // indirect
	github.com/yuin/goldmark v1.4.13 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/mod v0.9.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/term v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/tools v0.7.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
