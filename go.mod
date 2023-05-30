module github.com/uptrace/uptrace

go 1.20

replace github.com/segmentio/encoding => github.com/vmihailenco/encoding v0.3.4-0.20220121071420-f96fbbb25975

require (
	github.com/cespare/xxhash/v2 v2.2.0
	github.com/codemodus/kace v0.5.1
	github.com/gogo/protobuf v1.3.2
	github.com/klauspost/compress v1.16.5
	github.com/mileusna/useragent v1.3.2
	github.com/mostynb/go-grpc-compression v1.1.18
	github.com/rs/cors v1.9.0
	github.com/segmentio/encoding v0.3.6
	github.com/stretchr/testify v1.8.4
	github.com/uptrace/bunrouter v1.0.20
	github.com/uptrace/bunrouter/extra/bunrouterotel v1.0.20
	github.com/uptrace/bunrouter/extra/reqlog v1.0.20
	github.com/uptrace/go-clickhouse v0.3.2-0.20230530124502-03194bc43421
	github.com/uptrace/go-clickhouse/chdebug v0.3.2-0.20230530124502-03194bc43421
	github.com/uptrace/go-clickhouse/chotel v0.3.2-0.20230530124502-03194bc43421
	github.com/uptrace/opentelemetry-go-extra/otelzap v0.2.1
	github.com/uptrace/uptrace-go v1.16.0
	github.com/urfave/cli/v2 v2.25.5
	github.com/vmihailenco/msgpack/v5 v5.3.5
	github.com/vmihailenco/tagparser v0.1.2
	github.com/zyedidia/generic v1.2.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.42.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.42.0
	go.opentelemetry.io/otel v1.16.0
	go.opentelemetry.io/otel/metric v1.16.0
	go.opentelemetry.io/otel/trace v1.16.0
	go.opentelemetry.io/proto/otlp v0.19.0
	go.uber.org/zap v1.24.0
	go4.org v0.0.0-20230225012048-214862532bf5
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1
	google.golang.org/grpc v1.55.0
	google.golang.org/protobuf v1.30.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/cespare/xxhash v1.1.0
	github.com/coreos/go-oidc/v3 v3.6.0
	github.com/go-logr/zapr v1.2.4
	github.com/go-openapi/strfmt v0.21.7
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/prometheus/alertmanager v0.25.0
	github.com/slack-go/slack v0.12.2
	github.com/uptrace/bun/dialect/pgdialect v1.1.14
	github.com/uptrace/bun/driver/pgdriver v1.1.14
	github.com/vmihailenco/taskq/extra/oteltaskq/v4 v4.0.0-beta.4
	github.com/vmihailenco/taskq/pgq/v4 v4.0.0-beta.4
	github.com/vmihailenco/taskq/v4 v4.0.0-beta.4
	github.com/wneessen/go-mail v0.3.9
	github.com/zeebo/xxh3 v1.0.2
	golang.org/x/crypto v0.9.0
	golang.org/x/net v0.10.0
	golang.org/x/oauth2 v0.8.0
	gonum.org/v1/gonum v0.13.0
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-jose/go-jose/v3 v3.0.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.21.4 // indirect
	github.com/go-openapi/errors v0.20.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/loads v0.21.2 // indirect
	github.com/go-openapi/spec v0.20.9 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-openapi/validate v0.22.1 // indirect
	github.com/go-redis/redis_rate/v10 v10.0.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru v0.6.0 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/oklog/ulid/v2 v2.1.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/redis/go-redis/v9 v9.0.5 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/uptrace/opentelemetry-go-extra/otelsql v0.2.1 // indirect
	github.com/uptrace/opentelemetry-go-extra/otelutil v0.2.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.mongodb.org/mongo-driver v1.11.6 // indirect
	go.opentelemetry.io/contrib/instrumentation/runtime v0.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.16.0 // indirect
	go.opentelemetry.io/otel/sdk v1.16.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.39.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230526203410-71b5a4ffd15e // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230526203410-71b5a4ffd15e // indirect
	mellium.im/sasl v0.3.1 // indirect
)

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.15.2 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/uptrace/bun v1.1.14
	github.com/uptrace/bun/extra/bundebug v1.1.14
	github.com/uptrace/bun/extra/bunotel v1.1.14
	google.golang.org/genproto v0.0.0-20230526203410-71b5a4ffd15e // indirect
)
