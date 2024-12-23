module github.com/uptrace/pkg/idgen

go 1.23.3

replace github.com/segmentio/encoding => github.com/vmihailenco/encoding v0.3.4-0.20241127110735-a56c3dcf50f3

replace github.com/uptrace/pkg/clickhouse => ../clickhouse

replace github.com/uptrace/pkg/tagparser => ../tagparser

replace github.com/uptrace/pkg/msgp => ../msgp

replace github.com/uptrace/pkg/unsafeconv => ../unsafeconv

replace github.com/uptrace/pkg/unixtime => ../unixtime

require (
	github.com/segmentio/encoding v0.4.1
	github.com/stretchr/testify v1.10.0
	github.com/uptrace/bun v1.2.6
	github.com/uptrace/pkg/clickhouse v0.0.0-00010101000000-000000000000
	github.com/uptrace/pkg/msgp v0.0.0-00010101000000-000000000000
	github.com/uptrace/pkg/unsafeconv v0.0.0-00010101000000-000000000000
	github.com/zeebo/xxh3 v1.0.2
)

require (
	github.com/codemodus/kace v0.5.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/puzpuzpuz/xsync/v3 v3.4.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/uptrace/pkg/tagparser v0.0.0-00010101000000-000000000000 // indirect
	github.com/uptrace/pkg/unixtime v0.0.0-00010101000000-000000000000 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opentelemetry.io/otel v1.32.0 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	go.opentelemetry.io/otel/trace v1.32.0 // indirect
	golang.org/x/exp v0.0.0-20241204233417-43b7b7cde48d // indirect
	golang.org/x/sys v0.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
