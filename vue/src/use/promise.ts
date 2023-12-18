import { debounce } from 'lodash-es'
import axios from 'axios'
import { shallowRef, computed, watch } from 'vue'

// Composables
import { useSnackbar } from '@/use/snackbar'

// Misc
import { sentence } from '@/util/string'

type AsyncFunc = (...args: any[]) => Promise<any>

export interface Config {
  debounce?: number
  ignoreErrors?: boolean
}

export interface ApiError {
  code: string
  message: string

  statusCode: number
  traceId: string
  data: Record<string, any>
}

export function usePromise(fn: AsyncFunc, conf: Config = {}) {
  const snackbar = useSnackbar()

  const result = shallowRef<any>()
  const resultId = shallowRef(0)
  const rawError = shallowRef<any>()
  const status = shallowRef<Status>(Status.Unset)

  const pending = computed((): boolean => {
    switch (status.value) {
      case Status.Initing:
      case Status.Reloading:
        return true
    }
    return false
  })

  let currentId = 0

  let promised = (...args: any[]): Promise<any> => {
    switch (status.value) {
      case Status.Unset:
      case Status.Rejected:
        status.value = Status.Initing
        break
      case Status.Resolved:
        status.value = Status.Reloading
        break
    }

    let promise: Promise<any>

    currentId++
    ;(function (localID: number) {
      promise = fn(...args)
      promise.then(
        (res: any) => {
          if (localID === currentId) {
            resolve(res)
          }
        },
        (err: any) => {
          if (localID === currentId) {
            reject(err)
          }
          return err
        },
      )
    })(currentId)

    return promise
  }

  let resolve = (res: any): void => {
    result.value = res
    resultId.value = currentId
    rawError.value = undefined
    status.value = Status.Resolved
  }

  let reject = (err: any): void => {
    if (err === null || axios.isCancel(err)) {
      status.value = result.value !== undefined ? Status.Resolved : Status.Unset
      return
    }

    if (err === undefined) {
      result.value = undefined
      resultId.value = 0
      rawError.value = undefined
      status.value = Status.Unset
      return
    }

    result.value = undefined
    resultId.value = 0
    rawError.value = err
    status.value = Status.Rejected
  }

  let cancel = (): void => {
    if (status.value.pending()) {
      currentId++
    }
  }

  if (conf.debounce) {
    const debounced = debounce(promised, conf.debounce)

    const oldCancel = cancel
    cancel = () => {
      oldCancel()
      debounced.cancel()
    }

    const oldResolve = resolve
    const oldReject = reject

    promised = (...args: any[]): Promise<any> => {
      debounced(...args)
      return new Promise((promiseResolve, promiseReject) => {
        resolve = (res: any): void => {
          oldResolve(res)
          promiseResolve(res)
        }
        reject = (err: any): void => {
          oldReject(err)
          promiseReject(err)
        }
      })
    }
  }

  const error = computed((): ApiError | undefined => {
    const err = rawError.value
    if (!err) {
      return undefined
    }
    const data = err.response?.data ?? {}
    return {
      code: data?.code ?? '',
      message: errorMessage.value,

      statusCode: data?.statusCode ?? 0,
      traceId: data?.traceId ?? '',
      data,
    }
  })

  const errorMessage = computed((): string => {
    const err = rawError.value
    if (!err) {
      return ''
    }
    return sentence(err.response?.data?.error?.message ?? asString(err))
  })

  if (!conf.ignoreErrors) {
    watch(error, (error) => {
      if (!error) {
        return
      }
      switch (error.statusCode) {
        case 400:
        case 402:
        case 403:
          snackbar.notifyError(error.message)
      }
    })
  }

  return {
    status,
    pending,

    promised,
    result,
    resultId,
    rawError,
    error,
    errorMessage,

    cancel,
  }
}

export enum StatusValue {
  Unset = 'unset',
  Initing = 'initing',
  Resolved = 'resolved',
  Rejected = 'rejected',
  Reloading = 'reloading',
}

class Status {
  static Unset = new Status(StatusValue.Unset)
  static Initing = new Status(StatusValue.Initing)
  static Resolved = new Status(StatusValue.Resolved)
  static Rejected = new Status(StatusValue.Rejected)
  static Reloading = new Status(StatusValue.Reloading)

  value: StatusValue

  constructor(value: StatusValue) {
    this.value = value
  }

  toString(): string {
    return this.value
  }

  isUnset(): boolean {
    return this.value === StatusValue.Unset
  }

  initing(): boolean {
    return this.value === StatusValue.Initing
  }

  isResolved(): boolean {
    return this.value === StatusValue.Resolved
  }

  reloading(): boolean {
    return this.value === StatusValue.Reloading
  }

  isReady(): boolean {
    return !this.pending()
  }

  pending(): boolean {
    switch (this.value) {
      case StatusValue.Resolved:
      case StatusValue.Rejected:
        return false
      default:
        return true
    }
  }

  hasData(): boolean {
    switch (this.value) {
      case StatusValue.Resolved:
      case StatusValue.Reloading:
        return true
      default:
        return false
    }
  }
}

function asString(s: string | Error | undefined): string {
  if (!s) {
    return ''
  }
  if (typeof s === 'string') {
    return s
  }
  if (s.message) {
    return s.message
  }
  return ''
}
