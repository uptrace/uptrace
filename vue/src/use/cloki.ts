import { computed, proxyRefs } from '@vue/composition-api'

import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'

export type UseClokiSample = ReturnType<typeof useClokiSamples>

export interface Sample {
  string: string
  time: string
  labels: Record<string, any>
}

export function useClokiSamples(reqSource: AxiosRequestSource) {
  const { loading, data, reload } = useWatchAxios(() => {
    return reqSource()
  })

  const samples = computed((): Sample[] => {
    const samples = data.value?.samples ?? []
    return samples
  })

  return proxyRefs({
    loading,
    items: samples,

    reload,
  })
}
