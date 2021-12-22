import { useAxios } from '@/use/axios'
import {
  useWatchAxiosConfig,
  AxiosRequestSource,
  AxiosRequestConfig,
} from '@/use/watch-axios-config'

export type { AxiosRequestSource, AxiosRequestConfig }

export function useWatchAxios(source: AxiosRequestSource) {
  const {
    loading,
    data,
    error,

    request,
  } = useAxios()

  const { reload, abort } = useWatchAxiosConfig(
    source,
    (config, oldConfig, onInvalidate, abortCtrl) => {
      return request(config, abortCtrl)
    },
  )

  return {
    loading,

    data,
    error,

    abort,
    reload,
  }
}
