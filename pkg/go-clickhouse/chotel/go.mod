module github.com/uptrace/go-clickhouse/chotel

go 1.22

replace github.com/uptrace/go-clickhouse => ./..

replace github.com/uptrace/go-clickhouse/chdebug => ../chdebug

require (
	github.com/uptrace/go-clickhouse v0.3.1
	go.opentelemetry.io/otel v1.16.0
	go.opentelemetry.io/otel/trace v1.16.0
)

require (
	github.com/codemodus/kace v0.5.1 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	go.opentelemetry.io/otel/metric v1.16.0 // indirect
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
)
