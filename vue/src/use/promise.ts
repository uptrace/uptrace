import axios from 'axios'
import { shallowRef, computed, watch } from '@vue/composition-api'

// Composables
import { useSnackbar } from '@/use/snackbar'

type AsyncFunc = (...args: any[]) => Promise<any>

export function usePromise(fn: AsyncFunc) {
  const snackbar = useSnackbar()

  const result = shallowRef<any>()
  const error = shallowRef<any>()
  const pending = shallowRef(false)

  let id = 0

  let promised = (...args: any[]): Promise<any> => {
    let promise: Promise<any>

    id++
    ;(function (localID: number) {
      pending.value = true
      promise = fn(...args)
      promise.then(
        (res: any) => {
          pending.value = false
          if (localID === id) {
            resolve(res)
          }
        },
        (err: any) => {
          pending.value = false
          if (localID === id) {
            reject(err)
          }
          return err
        },
      )
    })(id)

    return promise
  }

  let resolve = (res: any): void => {
    result.value = res
    error.value = undefined
  }

  let reject = (err: any): void => {
    if (err === null || axios.isCancel(err)) {
      return
    }

    if (err === undefined) {
      result.value = undefined
      error.value = undefined
      return
    }

    result.value = undefined
    error.value = err
  }

  let cancel = (): void => {
    id++
  }

  const errorMessage = computed(() => {
    if (!error.value) {
      return ''
    }
    const msg = error.value.response?.data?.message
    if (msg) {
      return asString(msg)
    }
    return asString(error.value)
  })

  watch(error, (error) => {
    if (!error || !errorMessage.value) {
      return
    }
    switch (error.response?.status) {
      case 400:
      case 500:
        snackbar.notifyError(errorMessage.value)
    }
  })

  return {
    promised,
    result,
    error,
    pending,

    cancel,
  }
}

function asString(s: string | Error): string {
  if (typeof s === 'string') {
    return s
  }
  return s.message
}
