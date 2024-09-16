module github.com/uptrace/go-clickhouse

go 1.22

replace github.com/uptrace/go-clickhouse/chdebug => ./chdebug

require (
	github.com/bradleyjkemp/cupaloy v2.3.0+incompatible
	github.com/codemodus/kace v0.5.1
	github.com/jinzhu/inflection v1.0.0
	github.com/pierrec/lz4/v4 v4.1.21
	github.com/rs/dnscache v0.0.0-20230804202142-fc85eb664529
	github.com/stretchr/testify v1.8.4
	github.com/uptrace/go-clickhouse/chdebug v0.3.1
	go.opentelemetry.io/otel/trace v1.23.1
	golang.org/x/exp v0.0.0-20240213143201-ec583247a57a
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/otel v1.23.1 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
