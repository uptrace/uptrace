#!/usr/bin/env python
"""Django's command-line utility for administrative tasks."""
import os
import sys

from opentelemetry.instrumentation.logging import LoggingInstrumentor
from opentelemetry.instrumentation.django import DjangoInstrumentor
from opentelemetry.instrumentation.sqlite3 import SQLite3Instrumentor
import uptrace


def main():
    """Run administrative tasks."""
    os.environ.setdefault("DJANGO_SETTINGS_MODULE", "mysite.settings")

    uptrace.configure_opentelemetry(
        # Copy DSN here or use UPTRACE_DSN env var.
        # dsn="",
        service_name="myservice",
        service_version="v1.0.0",
    )

    LoggingInstrumentor().instrument(set_logging_format=True)
    SQLite3Instrumentor().instrument()
    DjangoInstrumentor().instrument()

    try:
        from django.core.management import execute_from_command_line
    except ImportError as exc:
        raise ImportError(
            "Couldn't import Django. Are you sure it's installed and "
            "available on your PYTHONPATH environment variable? Did you "
            "forget to activate a virtual environment?"
        ) from exc
    execute_from_command_line(sys.argv)


if __name__ == "__main__":
    main()
