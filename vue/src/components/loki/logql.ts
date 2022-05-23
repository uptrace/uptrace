import { computed, watch, proxyRefs } from '@vue/composition-api'

// Composables
import { usePager, PagerConfig } from '@/use/pager'
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'

export type UseLogql = ReturnType<typeof useLogql>

export interface Result {
  stream: Record<string, string>
  values: LogValue[]
}

export type LogValue = [string, string]

export interface LogqlConfig {
  pager?: PagerConfig
}

export function useLogql(reqSource: AxiosRequestSource, cfg: LogqlConfig = {}) {
  const pager = usePager(cfg.pager)

  const { loading, data } = useWatchAxios(reqSource)

  const results = computed((): Result[] => {
    return data.value?.data?.result ?? []
  })

  const numItem = computed(() => {
    return results.value.reduce((sum, result) => sum + result.values.length, 0)
  })

  watch(
    numItem,
    (numItem) => {
      pager.numItem = numItem
    },
    { immediate: true },
  )

  return proxyRefs({
    pager,

    loading,
    results,
    numItem,
  })
}

export function useLabels(reqSource: AxiosRequestSource) {
  const { loading, data } = useWatchAxios(reqSource)

  const labels = computed((): string[] => {
    return data.value?.data ?? []
  })

  return proxyRefs({ loading, items: labels })
}

export function useLabelValues(reqSource: AxiosRequestSource) {
  const { loading, data } = useWatchAxios(reqSource)

  const values = computed((): string[] => {
    return data.value?.data ?? []
  })

  return proxyRefs({ loading, items: values })
}
