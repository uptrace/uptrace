FROM alpine:latest as prep
RUN apk --update add ca-certificates
RUN apk add --no-cache curl && \
    curl -LJO https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.70.0/otelcol-contrib_0.70.0_linux_amd64.tar.gz && \
    tar -xzf otelcol-contrib_0.70.0_linux_amd64.tar.gz && \
    mv otelcol-contrib /otelcol-contrib && \
    chmod +x /otelcol-contrib && \
    apk del curl

COPY uptrace /
COPY entrypoint.sh /entrypoint.sh

EXPOSE 14317 14318 55679
ENTRYPOINT ["/entrypoint.sh"]
