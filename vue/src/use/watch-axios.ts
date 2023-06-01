import { cloneDeep } from 'lodash-es'
import { watch, proxyRefs, onBeforeUnmount } from 'vue'

import { useAxios, AxiosRequestConfig, AxiosConfig } from '@/use/axios'

export type AxiosRequest = AxiosRequestConfig | null | undefined
export type AxiosRequestSource = () => AxiosRequest

export type AxiosParams = Record<string, any> | null | undefined
export type AxiosParamsSource = () => AxiosParams

export interface AxiosWatchOptions extends AxiosConfig {
  immediate?: boolean
  notEqual?: boolean
  debounce?: number
  once?: boolean
}

export function watchAxios(source: AxiosRequestSource, options: AxiosWatchOptions = {}) {
  return proxyRefs(useWatchAxios(source, options))
}

export function useWatchAxios(source: AxiosRequestSource, options: AxiosWatchOptions = {}) {
  options.immediate = true
  if (options.debounce === undefined) {
    options.debounce = 10
  }
  if (options.notEqual === undefined) {
    options.notEqual = true
  }

  let abortCtrl: AbortController | undefined
  let prevReq: AxiosRequestConfig | undefined

  const {
    status,
    loading,
    data,
    error,
    errorCode,
    errorMessage,

    request,
  } = useAxios(options)

  const stopWatch = watch(
    source,
    (reqConf) => {
      if (reqConf === null) {
        if (abortCtrl) {
          abortCtrl.abort()
        }
        return
      }

      if (options.notEqual) {
        if (!axiosRequestChanged(reqConf, prevReq)) {
          return
        }
      }

      if (abortCtrl) {
        abortCtrl.abort()
      }
      makeRequest(reqConf)
    },
    { immediate: options.immediate },
  )
  onBeforeUnmount(abort)

  if (options.once) {
    const stopSelf = watch(status, (status) => {
      if (status.isResolved()) {
        stopWatch()
        stopSelf()
      }
    })
  }

  function reload() {
    if (!prevReq) {
      return Promise.reject()
    }

    if (abortCtrl) {
      return Promise.reject()
    }
    return makeRequest(prevReq)
  }

  function makeRequest(reqConf: AxiosRequestConfig | undefined) {
    abortCtrl = new AbortController()

    const promise = request(reqConf, abortCtrl).catch((err) => {
      if (err === undefined) {
        prevReq = undefined
      }
    })

    // Set prevReq immediately to ignore the reqConf flickering.
    prevReq = cloneDeep(reqConf)

    promise.finally(() => {
      abortCtrl = undefined
    })

    return promise
  }

  function abort() {
    if (abortCtrl) {
      abortCtrl.abort()
      abortCtrl = undefined
      prevReq = undefined
    }
  }

  return {
    status,
    loading,

    data,
    error,
    errorCode,
    errorMessage,

    abort,
    reload,
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
    if (key.startsWith(IGNORE_KEY_PREFIX)) {
      return undefined
    }
    if (ignoredKeys.indexOf(key) >= 0) {
      return undefined
    }
    if (value === undefined) {
      return null
    }
    return value
  })
}
