import { ref, proxyRefs } from '@vue/composition-api'

import { defineStore } from '@/use/store'

export const useSnackbar = defineStore('useSnackbar', () => {
  const notification = ref('')
  const active = ref(false)
  const color = ref('')
  const timeout = ref(-1)

  function notifySuccess(s: string) {
    notify(s)
    color.value = 'success'
  }

  function notifyError(s: string | Error) {
    notify(asString(s))
    color.value = 'error'
  }

  function notify(s: string) {
    notification.value = s
    active.value = true
    timeout.value = 10000
  }

  return proxyRefs({
    notification,
    active,
    color,
    timeout,

    notifySuccess,
    notifyError,
  })
})

function asString(s: string | Error): string {
  if (typeof s === 'string') {
    return s
  }
  return s.message
}
