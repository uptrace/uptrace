module github.com/uptrace/go-clickhouse/chotel

go 1.22

replace github.com/uptrace/go-clickhouse => ./..

replace github.com/uptrace/go-clickhouse/chdebug => ../chdebug

require (
	github.com/uptrace/go-clickhouse v0.3.1
	go.opentelemetry.io/otel v1.23.1
	go.opentelemetry.io/otel/trace v1.23.1
)

require (
	github.com/codemodus/kace v0.5.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/rs/dnscache v0.0.0-20230804202142-fc85eb664529 // indirect
	go.opentelemetry.io/otel/metric v1.23.1 // indirect
	golang.org/x/exp v0.0.0-20240213143201-ec583247a57a // indirect
	golang.org/x/sync v0.6.0 // indirect
)
