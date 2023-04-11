#!/bin/sh

getent passwd uptrace >/dev/null || useradd --system --user-group --no-create-home --shell /sbin/nologin uptrace

#install -d -m 0755 -o uptrace -g uptrace /var/lib/uptrace
