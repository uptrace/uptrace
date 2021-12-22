#!/bin/sh

set -euxo pipefail

/uptrace --config /etc/uptrace/uptrace.yml ch wait
/uptrace --config /etc/uptrace/uptrace.yml ch init
/uptrace --config /etc/uptrace/uptrace.yml ch migrate
/uptrace --config /etc/uptrace/uptrace.yml serve
