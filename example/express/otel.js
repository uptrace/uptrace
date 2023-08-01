'use strict'

const otel = require('@opentelemetry/api')
const { BatchSpanProcessor } = require('@opentelemetry/sdk-trace-base')
const { Resource } = require('@opentelemetry/resources')
const { NodeSDK } = require('@opentelemetry/sdk-node')
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-http')
const { AWSXRayIdGenerator } = require('@opentelemetry/id-generator-aws-xray')
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http')
const { ExpressInstrumentation } = require('@opentelemetry/instrumentation-express')

const dsn = process.env.UPTRACE_DSN || 'http://project2_secret_token@localhost:14318/2'
console.log('using dsn:', dsn)

const exporter = new OTLPTraceExporter({
  url: 'http://localhost:14318/v1/traces',
  headers: { 'uptrace-dsn': dsn },
  compression: 'gzip',
})
const bsp = new BatchSpanProcessor(exporter, {
  maxExportBatchSize: 1000,
  maxQueueSize: 1000,
})

const sdk = new NodeSDK({
  spanProcessor: bsp,
  resource: new Resource({
    'service.name': 'myservice',
    'service.version': '1.0.0',
  }),
  idGenerator: new AWSXRayIdGenerator(),
  instrumentations: [new HttpInstrumentation(), new ExpressInstrumentation()],
})
sdk.start()
