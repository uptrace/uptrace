import Vue from 'vue'
import * as Sentry from '@sentry/vue'

Sentry.init({
  app: Vue,
  dsn: 'http://project1_secret_token@localhost:14318/1',
  tracesSampleRate: 1.0,
})

// Set user information, as well as tags and further extras
Sentry.configureScope((scope) => {
  scope.setExtra('battery', 0.7)
  scope.setTag('user_mode', 'admin')
  scope.setUser({ id: '4711' })
  // scope.clear();
})

// Add a breadcrumb for future events
Sentry.addBreadcrumb({
  message: 'My Breadcrumb',
  // ...
})

// Capture exceptions, messages or manual events
Sentry.captureMessage('Hello, world!')
Sentry.captureException(new Error('Good bye'))
Sentry.captureEvent({
  message: 'Manual',
})
