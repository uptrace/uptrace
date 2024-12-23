module github.com/uptrace/pkg/msgp

go 1.23.3

replace github.com/segmentio/encoding => github.com/vmihailenco/encoding v0.3.4-0.20241127110735-a56c3dcf50f3

replace github.com/uptrace/pkg/tagparser => ../tagparser

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/google/go-cmp v0.6.0
	github.com/segmentio/encoding v0.4.1
	github.com/stretchr/testify v1.10.0
	github.com/uptrace/pkg/tagparser v0.0.0-00010101000000-000000000000
	github.com/vmihailenco/msgpack/v5 v5.4.1
	golang.org/x/tools v0.28.0
)

require (
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
