module github.com/uptrace/uptrace

go 1.18

replace github.com/segmentio/encoding => github.com/vmihailenco/encoding v0.3.4-0.20220121071420-f96fbbb25975

require (
	github.com/bradleyjkemp/cupaloy v2.3.0+incompatible
	github.com/cespare/xxhash/v2 v2.1.2
	github.com/codemodus/kace v0.5.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gogo/protobuf v1.3.2
	github.com/klauspost/compress v1.15.9
	github.com/mileusna/useragent v1.1.0
	github.com/mostynb/go-grpc-compression v1.1.16
	github.com/rs/cors v1.8.2
	github.com/segmentio/encoding v0.3.5
	github.com/stretchr/testify v1.8.0
	github.com/uptrace/bunrouter v1.0.19-0.20220828135510-86ff9250b2fc
	github.com/uptrace/bunrouter/extra/bunrouterotel v1.0.19-0.20220828135510-86ff9250b2fc
	github.com/uptrace/bunrouter/extra/reqlog v1.0.19-0.20220828135510-86ff9250b2fc
	github.com/uptrace/go-clickhouse v0.2.9-0.20220828135307-101b0312c333
	github.com/uptrace/go-clickhouse/chdebug v0.2.9-0.20220828135307-101b0312c333
	github.com/uptrace/go-clickhouse/chotel v0.2.9-0.20220828135307-101b0312c333
	github.com/uptrace/opentelemetry-go-extra/otelzap v0.1.15
	github.com/uptrace/uptrace-go v1.8.0
	github.com/urfave/cli/v2 v2.11.1
	github.com/vmihailenco/msgpack/v5 v5.3.5
	github.com/vmihailenco/tagparser v0.1.2
	github.com/zyedidia/generic v1.1.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.33.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.33.0
	go.opentelemetry.io/otel v1.9.0
	go.opentelemetry.io/otel/metric v0.31.0
	go.opentelemetry.io/otel/trace v1.9.0
	go.opentelemetry.io/proto/otlp v0.18.0
	go.uber.org/zap v1.21.0
	go4.org v0.0.0-20201209231011-d4a079459e60
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/go-openapi/strfmt v0.21.2
	github.com/prometheus/alertmanager v0.24.0
	golang.org/x/net v0.0.0-20220826154423-83b083e8dc8b
)

require (
	cloud.google.com/go/compute v1.7.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/coreos/go-oidc/v3 v3.4.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.21.2 // indirect
	github.com/go-openapi/errors v0.20.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/loads v0.21.1 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/go-openapi/validate v0.21.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-sqlite3 v1.14.15 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/uptrace/opentelemetry-go-extra/otelsql v0.1.15 // indirect
	github.com/uptrace/opentelemetry-go-extra/otelutil v0.1.15 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.mongodb.org/mongo-driver v1.8.3 // indirect
	go.opentelemetry.io/contrib/instrumentation/runtime v0.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.9.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.31.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.31.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.9.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.9.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.9.0 // indirect
	go.opentelemetry.io/otel/sdk v1.9.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.31.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/oauth2 v0.0.0-20220822191816-0ebed06d0094 // indirect
	golang.org/x/sys v0.0.0-20220825204002-c680a09ffe64 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.12 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.36.3 // indirect
	modernc.org/ccgo/v3 v3.16.9 // indirect
	modernc.org/libc v1.17.1 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.2.1 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/sqlite v1.18.1 // indirect
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.0.1 // indirect
)

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.1 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/uptrace/bun v1.1.8
	github.com/uptrace/bun/dialect/sqlitedialect v1.1.8
	github.com/uptrace/bun/driver/sqliteshim v1.1.8
	github.com/uptrace/bun/extra/bundebug v1.1.8
	github.com/uptrace/bun/extra/bunotel v1.1.8
	google.golang.org/genproto v0.0.0-20220802133213-ce4fa296bf78 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
