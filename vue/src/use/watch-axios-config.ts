import { AxiosRequestConfig, AxiosResponse } from 'axios'
import { watch, onBeforeUnmount } from 'vue'

export type { AxiosRequestConfig }

export type InvalidateCbRegistrator = (cb: () => void) => void

export type AxiosRequestSource = () => AxiosRequestConfig | null | undefined

export type AxiosWatchCallback = (
  config: AxiosRequestConfig | undefined,
  oldConfig: AxiosRequestConfig | undefined,
  onInvalidate: InvalidateCbRegistrator,
  abortCtrl: AbortController,
) => Promise<AxiosResponse<any>>

type VueCallBack<T> = (
  val: T | undefined,
  oldVal: T | undefined,
  onInvalidate: InvalidateCbRegistrator,
) => any

export function useWatchAxiosConfig(source: AxiosRequestSource, cb: AxiosWatchCallback) {
  let abortCtrl: AbortController | undefined

  let prevReq: AxiosRequestConfig | undefined
  let prevReqHash = ''

  let wrappedCB = (
    req: AxiosRequestConfig | undefined,
    _prevReq: AxiosRequestConfig | undefined,
    onInvalidate: InvalidateCbRegistrator,
  ) => {
    abortCtrl = new AbortController()
    cb(req, prevReq, onInvalidate, abortCtrl)
      .catch(() => {})
      .finally(() => {
        abortCtrl = undefined
      })
  }

  let handler: VueCallBack<AxiosRequestConfig> = (req, _prevReq, onInvalidate): void => {
    abort()

    wrappedCB(req, prevReq, onInvalidate)
    prevReq = req
    prevReqHash = JSON.stringify(prevReq)
  }

  const savedHandler = handler
  handler = function (req, prevReq, onInvalidate) {
    if (reqChanged(req)) {
      savedHandler(req, prevReq, onInvalidate)
    }
  }

  function abort() {
    if (!abortCtrl) {
      return
    }

    abortCtrl.abort()
    abortCtrl = undefined

    // Reset old request config since the request was canceled.
    prevReq = undefined
    prevReqHash = ''
  }

  function reload(req?: AxiosRequestConfig) {
    const onInvalidate = (): void => {}

    if (req) {
      if (abortCtrl && !reqChanged(req)) {
        return
      }
      return wrappedCB(req, prevReq, onInvalidate)
    }

    if (prevReq) {
      if (abortCtrl) {
        return
      }
      return wrappedCB(prevReq, undefined, onInvalidate)
    }
  }

  function reqChanged(req: AxiosRequestConfig | undefined) {
    return JSON.stringify(req) !== prevReqHash
  }

  const stopWatch = watch(source, handler, { immediate: true })

  onBeforeUnmount(abort)

  return {
    abort,
    reload,
    stopWatch,
  }
}
