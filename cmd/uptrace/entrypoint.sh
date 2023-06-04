#!/bin/sh

set -euxo pipefail

if [ $# -eq 0 ]; then
    /uptrace --config=/etc/uptrace/uptrace.yml pg wait
    /uptrace --config=/etc/uptrace/uptrace.yml ch wait
    /otelcol-contrib --config=/etc/otel/collector-config.yml &
    exec /uptrace --config=/etc/uptrace/uptrace.yml serve
else
    exec /uptrace --config=/etc/uptrace/uptrace.yml $@
fi
