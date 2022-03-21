#!/bin/sh

if command -v systemctl >/dev/null 2>&1; then
    systemctl stop uptrace.service
    systemctl disable uptrace.service
fi
