import axios, {
  AxiosRequestConfig as BaseAxiosRequestConfig,
  CancelTokenSource,
  AxiosResponse,
} from 'axios'
import { computed } from 'vue'

// Composables
import { usePromise, Config as PromiseConfig } from '@/use/promise'

export type AxiosConfig = PromiseConfig

export type AxiosParams = Record<string, string | undefined>

export interface AxiosRequestConfig extends BaseAxiosRequestConfig {
  ignore?: boolean
}

export function useAxios(cfg: AxiosConfig = {}) {
  let cancelToken: CancelTokenSource | null = null

  const {
    status,
    pending: loading,
    promised,
    result,
    error,
    errorMessage,
    cancel,
  } = usePromise((req: AxiosRequestConfig | null | undefined) => {
    if (req === null) {
      return Promise.reject(null)
    }
    if (!isValidReq(req)) {
      return Promise.reject(undefined)
    }
    if (req && req.ignore) {
      return Promise.reject(null)
    }

    if (req && !req.cancelToken) {
      cancelToken = axios.CancelToken.source()
      req = {
        ...req,
        cancelToken: cancelToken.token,
      }
    }

    return axios.request(req!)
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
    errorMessage,

    request,
    abort,
  }
}

function isValidReq(req: AxiosRequestConfig | undefined): boolean {
  if (req === undefined) {
    return false
  }

  if (req.url && req.url.includes('undefined')) {
    return false
  }

  if ('params' in req && req.params === undefined) {
    return false
  }
  if (req.params && !isValidData(req.params)) {
    return false
  }

  if ('data' in req && req.data === undefined) {
    return false
  }
  if (req.data && !isValidData(req.data)) {
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
