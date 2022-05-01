module github.com/uptrace/uptrace

go 1.18

replace github.com/segmentio/encoding => github.com/vmihailenco/encoding v0.3.4-0.20220121071420-f96fbbb25975

exclude go.opentelemetry.io/proto/otlp v0.15.0

require (
	github.com/bradleyjkemp/cupaloy v2.3.0+incompatible
	github.com/cespare/xxhash/v2 v2.1.2
	github.com/codemodus/kace v0.5.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/google/uuid v1.3.0
	github.com/klauspost/compress v1.15.1
	github.com/mileusna/useragent v1.0.2
	github.com/mostynb/go-grpc-compression v1.1.16
	github.com/rs/cors v1.8.2
	github.com/segmentio/encoding v0.3.4
	github.com/stretchr/testify v1.7.1
	github.com/uptrace/bunrouter v1.0.13
	github.com/uptrace/bunrouter/extra/bunrouterotel v1.0.13
	github.com/uptrace/bunrouter/extra/reqlog v1.0.13
	github.com/uptrace/go-clickhouse v0.2.6
	github.com/uptrace/go-clickhouse/chdebug v0.2.6
	github.com/uptrace/go-clickhouse/chotel v0.2.6
	github.com/uptrace/opentelemetry-go-extra/otelzap v0.1.11
	github.com/uptrace/uptrace-go v1.6.0
	github.com/urfave/cli/v2 v2.4.0
	github.com/vmihailenco/msgpack/v5 v5.3.5
	github.com/vmihailenco/tagparser v0.1.2
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.31.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.31.0
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
	go.opentelemetry.io/proto/otlp v0.14.0
	go.uber.org/zap v1.21.0
	go4.org v0.0.0-20201209231011-d4a079459e60
	golang.org/x/exp v0.0.0-20220428152302-39d4317da171
	google.golang.org/grpc v1.45.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/cenkalti/backoff/v4 v4.1.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.10.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pierrec/lz4/v4 v4.1.14 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/uptrace/opentelemetry-go-extra/otelutil v0.1.11 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/runtime v0.31.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.6.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.6.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.6.1 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.6.1 // indirect
	go.opentelemetry.io/otel/metric v0.28.0 // indirect
	go.opentelemetry.io/otel/sdk v1.6.1 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.28.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/net v0.0.0-20220325170049-de3da57026de // indirect
	golang.org/x/sys v0.0.0-20220429233432-b5fbb4746d32 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220328180837-c47567c462d1 // indirect
)
