import { cloneDeep } from 'lodash-es'
import { AxiosResponse } from 'axios'
import { watch, onBeforeUnmount, WatchOptions } from 'vue'

import { AxiosRequestConfig } from '@/use/axios'

export type { AxiosRequestConfig }

export type InvalidateCbRegistrator = (cb: () => void) => void

export type AxiosRequestSource = () => AxiosRequestConfig | null | undefined

export type AxiosParamsSource = () => Record<string, any>

export type AxiosWatchCallback = (
  config: AxiosRequestConfig | undefined,
  oldConfig: AxiosRequestConfig | undefined,
  onInvalidate: InvalidateCbRegistrator,
  abortCtrl: AbortController,
) => Promise<AxiosResponse<any>>

export interface AxiosWatchOptions extends WatchOptions {
  notEqual?: boolean
}

type VueCallBack<T> = (
  val: T | undefined,
  oldVal: T | undefined,
  onInvalidate: InvalidateCbRegistrator,
) => any

export function useWatchAxiosConfig(
  source: AxiosRequestSource,
  cb: AxiosWatchCallback,
  options: AxiosWatchOptions = {},
) {
  let abortCtrl: AbortController | undefined
  let prevReq: AxiosRequestConfig | undefined

  let wrappedCB = (
    req: AxiosRequestConfig | undefined,
    _prevReq: AxiosRequestConfig | undefined,
    onInvalidate: InvalidateCbRegistrator,
  ) => {
    if (abortCtrl) {
      abortCtrl.abort()
    }
    abortCtrl = new AbortController()

    const promise = cb(req, prevReq, onInvalidate, abortCtrl).catch(() => {})
    promise.finally(() => {
      abortCtrl = undefined
    })

    return promise
  }

  const handler: VueCallBack<AxiosRequestConfig> = (req, _prevReq, onInvalidate): void => {
    if (options.notEqual) {
      if (!axiosRequestChanged(req, prevReq)) {
        return
      }
    }

    wrappedCB(req, prevReq, onInvalidate)
    prevReq = cloneDeep(req)
  }

  const stopWatch = watch(source, handler, options)
  onBeforeUnmount(abort)

  function reload() {
    if (!prevReq) {
      return Promise.reject()
    }

    if (abortCtrl) {
      return Promise.reject()
    }
    return wrappedCB(prevReq, undefined, (): void => {})
  }

  function abort() {
    if (abortCtrl) {
      abortCtrl.abort()
      abortCtrl = undefined
      prevReq = undefined
    }
  }

  return {
    abort,
    reload,
    stopWatch,
  }
}

const IGNORE_KEY_PREFIX = '$ignore_'

function axiosRequestChanged(
  req: AxiosRequestConfig | undefined,
  prevReq: AxiosRequestConfig | undefined,
): boolean {
  const ignoredKeys: string[] = []

  if (req && req.params) {
    const keys = Object.keys(req.params)
      .filter((key) => key.startsWith(IGNORE_KEY_PREFIX))
      .map((key) => key.slice(IGNORE_KEY_PREFIX.length))
    ignoredKeys.push(...keys)
  }

  return hashAxiosRequest(req, ignoredKeys) != hashAxiosRequest(prevReq, ignoredKeys)
}

function hashAxiosRequest(req: AxiosRequestConfig | undefined, ignoredKeys: string[]): string {
  if (!req) {
    return ''
  }

  return JSON.stringify(req, (key: string, value: unknown): unknown => {
    if (ignoredKeys.indexOf(key) >= 0) {
      return undefined
    }
    if (key.startsWith(IGNORE_KEY_PREFIX)) {
      return undefined
    }
    if (value === undefined) {
      return null
    }
    return value
  })
}
