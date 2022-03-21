#!/bin/sh

if command -v systemctl >/dev/null 2>&1; then
    systemctl enable uptrace.service
    if [ -f /etc/uptrace/uptrace.yml ]; then
        systemctl start uptrace.service
    fi
fi
