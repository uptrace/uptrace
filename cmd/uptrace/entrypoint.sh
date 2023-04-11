#!/bin/sh

set -euxo pipefail

if [ $# -eq 0 ]; then
    /uptrace --config=/etc/uptrace/uptrace.yml pg wait
    /uptrace --config=/etc/uptrace/uptrace.yml ch wait
    exec /uptrace --config=/etc/uptrace/uptrace.yml serve
else
    exec /uptrace --config=/etc/uptrace/uptrace.yml $@
fi
