import { ref, proxyRefs, watch } from 'vue'

import { defineStore } from '@/use/store'

export const useSnackbar = defineStore(() => {
  const notification = ref('')
  const active = ref(false)
  const color = ref('')
  const timeout = ref(-1)
  const route = ref('')

  function notifySuccess(s: string) {
    notify(s)
    color.value = 'success'
  }

  function notifyError(s: string | Error) {
    notify(asString(s))
    color.value = 'error'
  }

  function notifyErrorWithDetails(s: string | Error, routeStr: string) {
    route.value = routeStr
    notifyError(s)
  }

  function notify(s: string) {
    notification.value = s
    active.value = true
    timeout.value = 10000
  }

  watch(active, (active) => {
    if (!active) {
      route.value = ''
    }
  })

  return proxyRefs({
    notification,
    active,
    color,
    timeout,
    route,

    notifySuccess,
    notifyError,
    notifyErrorWithDetails,
  })
})

function asString(s: string | Error): string {
  if (typeof s === 'string') {
    return s
  }
  return s.message
}
