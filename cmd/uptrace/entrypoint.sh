#!/bin/sh

set -euxo pipefail

if [ $# -eq 0 ]; then
    /uptrace --config=/etc/uptrace/uptrace.yml ch wait
    /uptrace --config=/etc/uptrace/uptrace.yml serve
else
    /uptrace --config=/etc/uptrace/uptrace.yml $@
fi
