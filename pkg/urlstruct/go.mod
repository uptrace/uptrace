module github.com/uptrace/pkg/urlstruct

go 1.23.3

replace github.com/segmentio/encoding => github.com/vmihailenco/encoding v0.3.4-0.20241127110735-a56c3dcf50f3

replace github.com/uptrace/pkg/unsafeconv => ../unsafeconv

replace github.com/uptrace/pkg/unixtime => ../unixtime

replace github.com/uptrace/pkg/tagparser => ../tagparser

replace github.com/uptrace/pkg/msgp => ../msgp

require (
	github.com/codemodus/kace v0.5.1
	github.com/puzpuzpuz/xsync/v3 v3.4.0
	github.com/segmentio/encoding v0.4.1
	github.com/stretchr/testify v1.10.0
	github.com/uptrace/pkg/unixtime v0.0.0-00010101000000-000000000000
	github.com/uptrace/uptrace v1.7.7
	github.com/vmihailenco/tagparser v0.1.2
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/uptrace/bun v1.1.18-0.20240129120547-ed6ed74d5379 // indirect
	github.com/uptrace/pkg/msgp v0.0.0-00010101000000-000000000000 // indirect
	github.com/uptrace/pkg/tagparser v0.0.0-00010101000000-000000000000 // indirect
	github.com/uptrace/pkg/unsafeconv v0.0.0-00010101000000-000000000000 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
