FROM alpine:latest as prep
RUN apk --update add ca-certificates

COPY uptrace /
COPY entrypoint.sh /entrypoint.sh

EXPOSE 4317 15678
ENTRYPOINT ["/entrypoint.sh"]
