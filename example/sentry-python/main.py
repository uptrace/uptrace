#!/usr/bin/env python3

import os

import sentry_sdk

dsn = os.environ.get("UPTRACE_DSN", "http://project1_secret@localhost:14318/1")
print("using DSN:", dsn)

sentry_sdk.init(
    dsn=dsn,
    # Set traces_sample_rate to 1.0 to capture 100%
    # of transactions for performance monitoring.
    # We recommend adjusting this value in production.
    traces_sample_rate=1.0,
)


def load_config():
    # The root cause: a low-level error deep in the call stack.
    config = {}
    return config["timeout"]


def connect():
    try:
        load_config()
    except KeyError as exc:
        # Re-raise as a higher-level error while preserving the original
        # cause. Sentry/Uptrace reports both exceptions as a linked chain.
        raise RuntimeError("failed to connect: invalid configuration") from exc


def main():
    try:
        connect()
    except RuntimeError:
        # Capture the full nested exception chain and report it to Uptrace.
        event_id = sentry_sdk.capture_exception()
        print("reported nested exception, event id:", event_id)


main()

# Ensure buffered events are delivered before the program exits.
sentry_sdk.flush()
