'using strict'

import { init, captureMessage } from '@sentry/browser'

init({
  dsn: 'http://project2_secret_token@localhost:14318/2',
  tracesSampleRate: 1.0,
})

const eventId = captureMessage('Hello, world!')
console.log("event id:", eventId)
