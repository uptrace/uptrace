import axios, { AxiosRequestConfig, CancelTokenSource, AxiosResponse } from 'axios'
import { computed } from 'vue'

// Composables
import { usePromise, Config } from '@/use/promise'

export type { Config }

export type AxiosParams = Record<string, string | undefined>

export function useAxios(cfg: Config = {}) {
  let cancelToken: CancelTokenSource | null = null

  const {
    status,
    pending: loading,
    promised,
    result,
    error,
    cancel,
  } = usePromise((config: AxiosRequestConfig | null | undefined) => {
    if (config === null) {
      return Promise.reject(null)
    }

    if (!isValidConfig(config)) {
      return Promise.reject(undefined)
    }

    if (config && !config.cancelToken) {
      cancelToken = axios.CancelToken.source()
      config = {
        ...config,
        cancelToken: cancelToken.token,
      }
    }

    return axios.request(config!)
  }, cfg)

  const data = computed(() => {
    return result.value?.data
  })

  function abort() {
    cancel()
    if (cancelToken) {
      cancelToken.cancel()
      cancelToken = null
    }
  }

  function request(
    config: AxiosRequestConfig | null | undefined,
    abortCtrl?: AbortController,
  ): Promise<AxiosResponse> {
    // TODO: this is unexpected and should be moved out of here
    abort()

    if (abortCtrl) {
      abortCtrl.signal.addEventListener('abort', () => {
        abort()
      })
    }

    return promised(config)
  }

  return {
    status,
    loading,

    result,
    data,
    error,

    request,
    abort,
  }
}

function isValidConfig(config: AxiosRequestConfig | undefined): boolean {
  if (config === undefined) {
    return false
  }

  if (config.url && config.url.includes('undefined')) {
    return false
  }

  if ('params' in config && config.params === undefined) {
    return false
  }
  if (config.params && !isValidData(config.params)) {
    return false
  }

  if ('data' in config && config.data === undefined) {
    return false
  }
  if (config.data && !isValidData(config.data)) {
    return false
  }

  return true
}

function isValidData(data: Record<string, any>): boolean {
  for (let key in data) {
    if (data[key] === undefined) {
      return false
    }
  }
  return true
}
