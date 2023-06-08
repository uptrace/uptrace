FROM alpine:latest as base

ARG TARGETPLATFORM

COPY bin/ /tmp/bin
RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2) && \
    mv /tmp/bin/uptrace_${GOOS}_${GOARCH} /uptrace

FROM alpine:latest

RUN apk --update add ca-certificates

COPY --from=base /uptrace .
COPY cmd/uptrace/entrypoint.sh /entrypoint.sh

EXPOSE 14317 14318
ENTRYPOINT ["/entrypoint.sh"]
