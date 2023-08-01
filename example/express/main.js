'use strict'

const port = 9999

const otel = require('@opentelemetry/api')
const express = require('express')
const app = express()
const tracer = otel.trace.getTracer('express-example')

app.get('/', indexHandler)
app.get('/hello/:username', helloHandler)

app.listen(9999, () => {
  console.log(`listening at http://localhost:${port}`)
})

function indexHandler(req, res) {
  const traceUrl = getTraceUrl(otel.trace.getSpan(otel.context.active()))
  res.send(
    `<html>` +
      `<p>Here are some routes for you:</p>` +
      `<ul>` +
      `<li><a href="/hello/world">Hello world</a></li>` +
      `<li><a href="/hello/foo-bar">Hello foo-bar</a></li>` +
      `</ul>` +
      `<p><a href="${traceUrl}">${traceUrl}</a></p>` +
      `</html>`,
  )
}

function helloHandler(req, res) {
  const span = trace.getSpan(otel.context.active())

  const err = new Error('User not found')
  span.recordException(err)

  const username = req.params.username
  const traceUrl = getTraceUrl(span)
  res.send(
    `<html>` +
      `<h3>Hello ${username}</h3>` +
      `<p><a href="${traceUrl}">${traceUrl}</a></p>` +
      `</html>`,
  )
}

function getTraceUrl(span) {
  const ctx = span.spanContext()
  const traceId = ctx?.traceId ?? '<no span>'
  return `http://localhost:14318/traces/${traceId}`
}
