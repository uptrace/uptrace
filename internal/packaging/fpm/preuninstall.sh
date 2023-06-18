#!/bin/sh

if command -v systemctl >/dev/null 2>&1; then
    systemctl stop uptrace.service || true
    systemctl disable uptrace.service || true
fi
