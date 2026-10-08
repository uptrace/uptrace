# OpenTelemetry Express example for Uptrace

Install dependencies:

```bash
npm install
```

Start Express server:

```bash
UPTRACE_DSN="http://project1_secret@localhost:14318?grpc=14317" node --require ./otel.js main.js
```

Then open http://localhost:9999

See [OpenTelemetry Express.js](https://uptrace.dev/get/instrument/opentelemetry-express.html) for
details.
