import { useAxios, Config } from '@/use/axios'
import {
  useWatchAxiosConfig,
  AxiosRequestSource,
  AxiosRequestConfig,
} from '@/use/watch-axios-config'

export type { AxiosRequestSource, AxiosRequestConfig }

export function useWatchAxios(source: AxiosRequestSource, cfg: Config = {}) {
  const {
    status,
    loading,
    data,
    error,

    request,
  } = useAxios(cfg)

  const { reload, abort } = useWatchAxiosConfig(
    source,
    (config, oldConfig, onInvalidate, abortCtrl) => {
      return request(config, abortCtrl)
    },
  )

  return {
    status,
    loading,

    data,
    error,

    abort,
    reload,
  }
}
