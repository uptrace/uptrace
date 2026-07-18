'use strict'

const express = require('express')
const otel = require('@opentelemetry/api')

const app = express()
const port = 9999

app.get('/', indexHandler)
app.get('/hello/:username', helloHandler)

app.listen(port, () => {
  console.log(`listening at http://localhost:${port}`)
})

function indexHandler(req, res) {
  const span = getActiveSpan()
  const traceUrl = getTraceUrl(span)

  res.send(`
    <html>
      <p>Available routes:</p>
      <ul>
        <li><a href="/hello/world">Hello world</a></li>
        <li><a href="/hello/foo-bar">Hello foo-bar</a></li>
      </ul>
      <p>
        <a href="${traceUrl}">${traceUrl}</a>
      </p>
    </html>
  `)
}

function helloHandler(req, res) {
  const span = getActiveSpan()

  const username = req.params.username
  const traceUrl = getTraceUrl(span)

  res.send(`
    <html>
      <h3>Hello ${username}</h3>
      <p>
        <a href="${traceUrl}">${traceUrl}</a>
      </p>
    </html>
  `)
}

function getActiveSpan() {
  return otel.trace.getSpan(otel.context.active())
}

function getTraceUrl(span) {
  if (!span) {
    return 'No active trace'
  }

  const { traceId } = span.spanContext()

  return `http://localhost:14318/traces/${traceId}`
}