module github.com/uptrace/pkg/clickhouse

go 1.23.3

replace github.com/segmentio/encoding => github.com/vmihailenco/encoding v0.3.4-0.20241127110735-a56c3dcf50f3

replace github.com/uptrace/pkg/unsafeconv => ../unsafeconv

replace github.com/uptrace/pkg/tagparser => ../tagparser

replace github.com/uptrace/pkg/msgp => ../msgp

replace github.com/uptrace/pkg/unixtime => ../unixtime

require (
	github.com/Masterminds/sprig/v3 v3.3.0
	github.com/bradleyjkemp/cupaloy v2.3.0+incompatible
	github.com/brianvoe/gofakeit/v5 v5.11.2
	github.com/codemodus/kace v0.5.1
	github.com/fatih/color v1.18.0
	github.com/go-faster/city v1.0.1
	github.com/jinzhu/inflection v1.0.0
	github.com/klauspost/compress v1.17.11
	github.com/pierrec/lz4/v4 v4.1.21
	github.com/puzpuzpuz/xsync/v3 v3.4.0
	github.com/segmentio/asm v1.2.0
	github.com/segmentio/encoding v0.4.1
	github.com/stretchr/testify v1.10.0
	github.com/uptrace/pkg/msgp v0.0.0-00010101000000-000000000000
	github.com/uptrace/pkg/tagparser v0.0.0-00010101000000-000000000000
	github.com/uptrace/pkg/unixtime v0.0.0-00010101000000-000000000000
	github.com/uptrace/pkg/unsafeconv v0.0.0-00010101000000-000000000000
	github.com/zeebo/xxh3 v1.0.2
	go.opentelemetry.io/otel v1.32.0
	go.opentelemetry.io/otel/trace v1.32.0
	go4.org v0.0.0-20230225012048-214862532bf5
	golang.org/x/exp v0.0.0-20241204233417-43b7b7cde48d
)

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	golang.org/x/crypto v0.30.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
