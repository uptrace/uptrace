import axios from 'axios'
import { shallowRef, computed, watch } from 'vue'

// Composables
import { useSnackbar } from '@/use/snackbar'

type AsyncFunc = (...args: any[]) => Promise<any>

export interface Config {
  ignoreErrors?: boolean
}

export enum StatusValue {
  Unset = 'unset',
  Initing = 'initing',
  Resolved = 'resolved',
  Rejected = 'rejected',
  Reloading = 'reloading',
}

export function usePromise(fn: AsyncFunc, cfg: Config = {}) {
  const snackbar = useSnackbar()

  const result = shallowRef<any>()
  const error = shallowRef<any>()
  const status = shallowRef<Status>(Status.Unset)

  const pending = computed((): boolean => {
    switch (status.value) {
      case Status.Initing:
      case Status.Reloading:
        return true
    }
    return false
  })

  let id = 0

  let promised = (...args: any[]): Promise<any> => {
    switch (status.value) {
      case Status.Unset:
        status.value = Status.Initing
        break
      case Status.Resolved:
        status.value = Status.Reloading
        break
    }

    let promise: Promise<any>

    id++
    ;(function (localID: number) {
      promise = fn(...args)
      promise.then(
        (res: any) => {
          if (localID === id) {
            resolve(res)
          }
        },
        (err: any) => {
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
    status.value = Status.Resolved
  }

  let reject = (err: any): void => {
    if (err === null || axios.isCancel(err)) {
      status.value = result.value !== undefined ? Status.Resolved : Status.Unset
      return
    }

    if (err === undefined) {
      result.value = undefined
      error.value = undefined
      status.value = Status.Unset
      return
    }

    result.value = undefined
    error.value = err
    status.value = Status.Rejected
  }

  let cancel = (): void => {
    id++
  }

  const errorMessage = computed(() => {
    const msg = error.value?.response?.data?.message
    if (msg) {
      return msg
    }
    return asString(error.value)
  })

  if (!cfg.ignoreErrors) {
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
  }

  return {
    status,
    pending,

    promised,
    result,
    error,
    errorMessage,

    cancel,
  }
}

//------------------------------------------------------------------------------

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

  hasData(): boolean {
    switch (this.value) {
      case StatusValue.Resolved:
      case StatusValue.Reloading:
        return true
    }
    return false
  }
}

function asString(s: string | Error): string {
  if (typeof s === 'string') {
    return s
  }
  return s.message
}
