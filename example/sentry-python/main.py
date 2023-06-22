#!/usr/bin/env python3

import os

import sentry_sdk

dsn = os.environ.get("UPTRACE_DSN", "http://project2_secret_token@localhost:14318/2")
print("using DSN:", dsn)

sentry_sdk.init(
    dsn=dsn,
    # Set traces_sample_rate to 1.0 to capture 100%
    # of transactions for performance monitoring.
    # We recommend adjusting this value in production.
    traces_sample_rate=1.0,
)

division_by_zero = 1 / 0
