FROM alpine:latest as prep
RUN apk --update add ca-certificates

COPY uptrace /
COPY entrypoint.sh /entrypoint.sh

EXPOSE 14317 14318
ENTRYPOINT ["/entrypoint.sh"]
