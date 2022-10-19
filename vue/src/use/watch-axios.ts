import { watch, proxyRefs } from 'vue'

import { useAxios, AxiosConfig } from '@/use/axios'
import {
  useWatchAxiosConfig,
  AxiosRequestSource,
  AxiosParamsSource,
  AxiosWatchOptions as BaseAxiosWatchOptions,
  AxiosRequestConfig,
} from '@/use/watch-axios-config'

export type { AxiosRequestSource, AxiosParamsSource, AxiosRequestConfig }

export interface AxiosWatchOptions extends BaseAxiosWatchOptions, AxiosConfig {
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

  const {
    status,
    loading,
    data,
    error,
    errorMessage,

    request,
  } = useAxios({ debounce: options.debounce })

  const { reload, abort, stopWatch } = useWatchAxiosConfig(
    source,
    (config, oldConfig, onInvalidate, abortCtrl) => {
      return request(config!, abortCtrl)
    },
    options,
  )

  if (options.once) {
    const stopSelf = watch(status, (status) => {
      if (status.isResolved()) {
        stopWatch()
        stopSelf()
      }
    })
  }

  return {
    status,
    loading,

    data,
    error,
    errorMessage,

    abort,
    reload,
  }
}
