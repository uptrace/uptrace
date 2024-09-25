module github.com/uptrace/uptrace

go 1.22

replace github.com/segmentio/encoding => github.com/vmihailenco/encoding v0.3.4-0.20240424103427-05bd731dd667

replace github.com/grafana/tempo => ./pkg/tempo

replace github.com/uptrace/go-clickhouse => ./pkg/go-clickhouse

replace github.com/vmihailenco/taskq/v4 => ./pkg/taskq

replace github.com/vmihailenco/taskq/pgq/v4 => ./pkg/taskq/pgq

replace github.com/vmihailenco/taskq/extra/oteltaskq/v4 => ./pkg/taskq/extra/oteltaskq

require (
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/codemodus/kace v0.5.1
	github.com/coreos/go-oidc/v3 v3.9.0
	github.com/go-logr/zapr v1.3.0
	github.com/go-openapi/strfmt v0.22.0
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/grafana/tempo v1.5.0
	github.com/klauspost/compress v1.17.6
	github.com/mileusna/useragent v1.3.4
	github.com/mostynb/go-grpc-compression v1.2.2
	github.com/prometheus/alertmanager v0.26.0
	github.com/prometheus/prometheus v0.49.1
	github.com/rs/cors v1.10.1
	github.com/segmentio/encoding v0.4.0
	github.com/slack-go/slack v0.12.4
	github.com/stretchr/testify v1.9.0
	github.com/uptrace/bun v1.1.18-0.20240129120547-ed6ed74d5379
	github.com/uptrace/bun/dialect/pgdialect v1.1.18-0.20240129120547-ed6ed74d5379
	github.com/uptrace/bun/driver/pgdriver v1.1.18-0.20240129120547-ed6ed74d5379
	github.com/uptrace/bun/extra/bundebug v1.1.18-0.20240129120547-ed6ed74d5379
	github.com/uptrace/bun/extra/bunotel v1.1.18-0.20240129120547-ed6ed74d5379
	github.com/uptrace/bunrouter v1.0.21
	github.com/uptrace/bunrouter/extra/bunrouterotel v1.0.21
	github.com/uptrace/bunrouter/extra/reqlog v1.0.21
	github.com/uptrace/go-clickhouse v0.3.1
	github.com/uptrace/go-clickhouse/chdebug v0.3.1
	github.com/uptrace/go-clickhouse/chotel v0.3.1
	github.com/uptrace/opentelemetry-go-extra/otelzap v0.2.3
	github.com/uptrace/uptrace-go v1.28.0
	github.com/urfave/cli/v2 v2.27.1
	github.com/vmihailenco/msgpack/v5 v5.4.1
	github.com/vmihailenco/tagparser v0.1.2
	github.com/vmihailenco/taskq/extra/oteltaskq/v4 v4.0.0-beta.4
	github.com/vmihailenco/taskq/pgq/v4 v4.0.0-beta.4
	github.com/vmihailenco/taskq/v4 v4.0.0-beta.4
	github.com/wk8/go-ordered-map/v2 v2.1.8
	github.com/wneessen/go-mail v0.4.3
	github.com/xhit/go-str2duration/v2 v2.1.0
	github.com/zeebo/xxh3 v1.0.2
	github.com/zyedidia/generic v1.2.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.48.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.48.0
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/metric v1.28.0
	go.opentelemetry.io/otel/trace v1.28.0
	go.opentelemetry.io/proto/otlp v1.3.1
	go.uber.org/zap v1.26.0
	go4.org v0.0.0-20230225012048-214862532bf5
	golang.org/x/crypto v0.25.0
	golang.org/x/exp v0.0.0-20240213143201-ec583247a57a
	golang.org/x/net v0.27.0
	golang.org/x/oauth2 v0.20.0
	gonum.org/v1/gonum v0.14.0
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.2
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.9.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.5.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.5.2 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.2 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/alecthomas/units v0.0.0-20231202071711-9a357b53e9c9 // indirect
	github.com/apache/thrift v0.19.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go v1.50.20 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/edsrzf/mmap-go v1.1.0 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-jose/go-jose/v3 v3.0.1 // indirect
	github.com/go-kit/kit v0.13.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.22.2 // indirect
	github.com/go-openapi/errors v0.21.0 // indirect
	github.com/go-openapi/jsonpointer v0.20.2 // indirect
	github.com/go-openapi/jsonreference v0.20.4 // indirect
	github.com/go-openapi/loads v0.21.5 // indirect
	github.com/go-openapi/spec v0.20.14 // indirect
	github.com/go-openapi/swag v0.22.9 // indirect
	github.com/go-openapi/validate v0.23.0 // indirect
	github.com/go-redis/redis_rate/v10 v10.0.1 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/status v1.1.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/handlers v1.5.2 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/grafana/dskit v0.0.0-20240216094944-102ab37233e4 // indirect
	github.com/grafana/pyroscope-go/godeltaprof v0.1.7 // indirect
	github.com/grafana/regexp v0.0.0-20221123153739-15dc172cd2db // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jaegertracing/jaeger v1.54.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/oklog/ulid/v2 v2.1.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.94.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger v0.94.0 // indirect
	github.com/opentracing-contrib/go-grpc v0.0.0-20210225150812-73cb765af46e // indirect
	github.com/opentracing-contrib/go-stdlib v1.0.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.1.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.18.0 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/prometheus/common v0.47.0 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/exporter-toolkit v0.11.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/redis/go-redis/v9 v9.5.0 // indirect
	github.com/rs/dnscache v0.0.0-20230804202142-fc85eb664529 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/sercand/kuberesolver/v5 v5.1.1 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/cobra v1.8.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.18.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/uptrace/opentelemetry-go-extra/otelsql v0.2.3 // indirect
	github.com/uptrace/opentelemetry-go-extra/otelutil v0.2.3 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xrash/smetrics v0.0.0-20231213231151-1d8dd44e695e // indirect
	go.mongodb.org/mongo-driver v1.14.0 // indirect
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
	go.uber.org/goleak v1.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240723171418-e6d459c13d2a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240723171418-e6d459c13d2a // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	mellium.im/sasl v0.3.1 // indirect
)