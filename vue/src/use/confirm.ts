import { shallowRef } from 'vue'

import { defineStore } from '@/use/store'

export const useConfirm = defineStore(() => {
  const dialog = shallowRef(false)
  const message = shallowRef('')
  const title = shallowRef('')
  const width = shallowRef(400)

  let promiseResolve: (value: unknown) => void
  let promiseReject: () => void

  function open(titleValue: string, msg: string) {
    title.value = titleValue
    message.value = msg
    dialog.value = true

    return new Promise((resolve, reject) => {
      promiseResolve = resolve
      promiseReject = reject
    })
  }

  function agree() {
    promiseResolve(undefined)
    dialog.value = false
  }

  function cancel() {
    promiseReject()
    dialog.value = false
  }

  return {
    dialog,
    message,
    title,
    width,

    open,
    agree,
    cancel,
  }
})
