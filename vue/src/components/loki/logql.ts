import { computed, watch, proxyRefs } from '@vue/composition-api'

// Composables
import { usePager, PagerConfig } from '@/use/pager'
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'

export type UseLogql = ReturnType<typeof useLogql>

interface Result {
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

  const result = computed((): Result | undefined => {
    const result = data.value?.data?.result ?? []
    if (result.length) {
      return result[0]
    }
    return undefined
  })

  const labels = computed((): Record<string, string> => {
    return result.value?.stream ?? {}
  })

  const logs = computed((): LogValue[] => {
    return result.value?.values ?? []
  })

  watch(
    () => logs.value.length,
    (numItem) => {
      pager.numItem = numItem
    },
    { immediate: true },
  )

  return proxyRefs({
    pager,

    loading,
    labels,
    logs,
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
