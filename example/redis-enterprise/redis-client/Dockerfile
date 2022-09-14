FROM golang:alpine AS base
COPY . /src/
WORKDIR /src/redis-client

FROM base AS redis-client
RUN go install ./client.go
CMD ["/go/bin/client"]
