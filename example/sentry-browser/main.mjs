'using strict'

import { init, captureMessage } from '@sentry/browser'

let dsn = process.env.UPTRACE_DSN
if (!dsn) {
  dsn = 'http://project2_secret_token@localhost:14318/2'
}
console.log('using dsn:', dsn)

init({
  dsn: dsn,
  tracesSampleRate: 1.0,
})

const eventId = captureMessage('Hello, world!')
console.log("event id:", eventId)
