#!/bin/sh

getent passwd uptrace >/dev/null || useradd --system --user-group --no-create-home --shell /sbin/nologin uptrace
