import { computed, proxyRefs } from '@vue/composition-api'

// Composables
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'

export type UsePercentiles = ReturnType<typeof usePercentiles>

export function usePercentiles(axiosSource: AxiosRequestSource) {
  const { loading, data } = useWatchAxios(axiosSource)

  const count = computed((): number => {
    return data.value?.count ?? 0
  })

  const rate = computed((): number => {
    return data.value?.rate ?? 0
  })

  const errorCount = computed((): number => {
    return data.value?.errorCount ?? 0
  })

  const p50 = computed((): number => {
    return data.value?.p50 ?? 0
  })

  const p90 = computed((): number => {
    return data.value?.p90 ?? 0
  })

  const max = computed((): number => {
    return data.value?.max ?? 0
  })

  return proxyRefs({
    loading,
    data,

    count,
    rate,
    errorCount,
    p50,
    p90,
    max,
  })
}
