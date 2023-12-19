import { computed, proxyRefs } from 'vue'

// Composables
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'

export type UsePercentiles = ReturnType<typeof usePercentiles>

export function usePercentiles(axiosSource: AxiosRequestSource) {
  const { status, loading, data } = useWatchAxios(axiosSource)

  const stats = computed(() => {
    return data.value?.stats ?? {}
  })

  return proxyRefs({
    status,
    loading,

    stats,
  })
}
