# OpenTelemetry Express example for Uptrace

This example demonstrates manual OpenTelemetry configuration.

For a simpler Uptrace setup, see the [@uptrace/node](https://npmjs.com/package/@uptrace/node) SDK.

Install dependencies:

```bash
npm install
```

Start Express server:

```bash
npm start
```

To send traces to Uptrace, set the `UPTRACE_DSN` environment variable:

```bash
UPTRACE_DSN="http://project1_secret@localhost:14318" npm start
```

Then open http://localhost:9999

See [OpenTelemetry Express.js](https://uptrace.dev/get/instrument/opentelemetry-express.html) for
details.
