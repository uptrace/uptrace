module github.com/uptrace/uptrace

go 1.23.3

replace github.com/segmentio/encoding => github.com/vmihailenco/encoding v0.3.4-0.20241127110735-a56c3dcf50f3

replace github.com/grafana/tempo => ./pkg/tempo

replace github.com/uptrace/pkg/clickhouse => ./pkg/clickhouse

replace github.com/uptrace/pkg/unsafeconv => ./pkg/unsafeconv

replace github.com/uptrace/pkg/tagparser => ./pkg/tagparser

replace github.com/uptrace/pkg/msgp => ./pkg/msgp

replace github.com/uptrace/pkg/unixtime => ./pkg/unixtime

replace github.com/uptrace/pkg/idgen => ./pkg/idgen

replace github.com/uptrace/pkg/urlstruct => ./pkg/urlstruct

replace github.com/vmihailenco/taskq/v4 => ./pkg/taskq

replace github.com/vmihailenco/taskq/pgq/v4 => ./pkg/taskq/pgq

replace github.com/vmihailenco/taskq/extra/oteltaskq/v4 => ./pkg/taskq/extra/oteltaskq

require (
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/coreos/go-oidc/v3 v3.9.0
	github.com/go-logr/zapr v1.3.0
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/klauspost/compress v1.18.0
	github.com/mileusna/useragent v1.3.4
	github.com/mostynb/go-grpc-compression v1.2.2
	github.com/prometheus/prometheus v0.49.1
	github.com/rs/cors v1.10.1
	github.com/segmentio/encoding v0.4.1
	github.com/stretchr/testify v1.10.0
	github.com/uptrace/bun v1.2.7-0.20241126124946-928d0779110e
	github.com/uptrace/bun/dialect/pgdialect v1.2.7-0.20241126124946-928d0779110e
	github.com/uptrace/bun/driver/pgdriver v1.2.7-0.20241126124946-928d0779110e
	github.com/uptrace/bun/extra/bundebug v1.2.7-0.20241126124946-928d0779110e
	github.com/uptrace/bun/extra/bunotel v1.2.7-0.20241126124946-928d0779110e
	github.com/uptrace/bunrouter v1.0.22
	github.com/uptrace/bunrouter/extra/bunrouterotel v1.0.22
	github.com/uptrace/bunrouter/extra/reqlog v1.0.22
	github.com/uptrace/opentelemetry-go-extra/otelzap v0.2.3
	github.com/uptrace/pkg/clickhouse v0.0.0-00010101000000-000000000000
	github.com/uptrace/pkg/idgen v0.0.0-00010101000000-000000000000
	github.com/uptrace/pkg/unixtime v0.0.0-00010101000000-000000000000
	github.com/uptrace/pkg/unsafeconv v0.0.0-00010101000000-000000000000
	github.com/uptrace/pkg/urlstruct v0.0.0-00010101000000-000000000000
	github.com/uptrace/uptrace-go v1.28.0
	github.com/urfave/cli/v2 v2.27.7
	github.com/vmihailenco/msgpack/v5 v5.4.1
	github.com/vmihailenco/taskq/extra/oteltaskq/v4 v4.0.0-beta.4
	github.com/vmihailenco/taskq/pgq/v4 v4.0.0-beta.4
	github.com/vmihailenco/taskq/v4 v4.0.0-beta.4
	github.com/wk8/go-ordered-map/v2 v2.1.8
	github.com/wneessen/go-mail v0.4.3
	github.com/xhit/go-str2duration/v2 v2.1.0
	github.com/zyedidia/generic v1.2.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.48.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.48.0
	go.opentelemetry.io/otel v1.32.0
	go.opentelemetry.io/otel/metric v1.32.0
	go.opentelemetry.io/otel/trace v1.32.0
	go.opentelemetry.io/proto/otlp v1.3.1
	go.uber.org/fx v1.23.0
	go.uber.org/zap v1.26.0
	go4.org v0.0.0-20230225012048-214862532bf5
	golang.org/x/crypto v0.40.0
	golang.org/x/exp v0.0.0-20241204233417-43b7b7cde48d
	golang.org/x/net v0.41.0
	gonum.org/v1/gonum v0.14.0
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.2
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.9.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.5.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.5.2 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.2 // indirect
	github.com/alecthomas/units v0.0.0-20231202071711-9a357b53e9c9 // indirect
	github.com/aws/aws-sdk-go v1.50.20 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/codemodus/kace v0.5.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-jose/go-jose/v3 v3.0.1 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonreference v0.20.4 // indirect
	github.com/go-openapi/swag v0.22.9 // indirect
	github.com/go-redis/redis_rate/v10 v10.0.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/grafana/regexp v0.0.0-20221123153739-15dc172cd2db // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/hashicorp/go-hclog v1.6.2 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/oklog/ulid/v2 v2.1.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc5 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.18.0 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/prometheus/common v0.47.0 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/puzpuzpuz/xsync/v3 v3.4.0 // indirect
	github.com/redis/go-redis/v9 v9.5.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/uptrace/opentelemetry-go-extra/otelsql v0.3.2 // indirect
	github.com/uptrace/opentelemetry-go-extra/otelutil v0.2.3 // indirect
	github.com/uptrace/pkg/msgp v0.0.0-00010101000000-000000000000 // indirect
	github.com/uptrace/pkg/tagparser v0.0.0-00010101000000-000000000000 // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/collector/featuregate v1.1.0 // indirect
	go.opentelemetry.io/collector/pdata v1.1.0 // indirect
	go.opentelemetry.io/collector/semconv v0.94.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/runtime v0.53.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.4.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.28.0 // indirect
	go.opentelemetry.io/otel/log v0.4.0 // indirect
	go.opentelemetry.io/otel/sdk v1.28.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.4.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.28.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/dig v1.18.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/oauth2 v0.20.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240723171418-e6d459c13d2a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240723171418-e6d459c13d2a // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	mellium.im/sasl v0.3.2 // indirect
)
