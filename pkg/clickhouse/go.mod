module github.com/uptrace/pkg/clickhouse

go 1.23.3

replace github.com/segmentio/encoding => github.com/vmihailenco/encoding v0.3.4-0.20241226104945-d6ad09eb6898

replace github.com/uptrace/pkg/unsafeconv => ../unsafeconv

replace github.com/uptrace/pkg/tagparser => ../tagparser

replace github.com/uptrace/pkg/msgp => ../msgp

replace github.com/uptrace/pkg/unixtime => ../unixtime

require (
	github.com/codemodus/kace v0.5.1
	github.com/fatih/color v1.18.0
	github.com/go-faster/city v1.0.1
	github.com/jinzhu/inflection v1.0.0
	github.com/klauspost/compress v1.17.11
	github.com/pierrec/lz4/v4 v4.1.21
	github.com/puzpuzpuz/xsync/v3 v3.4.0
	github.com/segmentio/asm v1.2.0
	github.com/segmentio/encoding v0.4.1
	github.com/uptrace/pkg/msgp v0.0.0-00010101000000-000000000000
	github.com/uptrace/pkg/tagparser v0.0.0-00010101000000-000000000000
	github.com/uptrace/pkg/unixtime v0.0.0-00010101000000-000000000000
	github.com/uptrace/pkg/unsafeconv v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.32.0
	go.opentelemetry.io/otel/trace v1.32.0
	golang.org/x/exp v0.0.0-20241204233417-43b7b7cde48d
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)
