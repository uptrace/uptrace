'using strict'

import { init, captureMessage } from '@sentry/browser'

init({
  dsn: 'http://project2_secret_token@localhost:14318/2',
  tracesSampleRate: 1.0,
  debug: true,
})

captureMessage('Hello, world!')
