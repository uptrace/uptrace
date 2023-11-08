import { computed, proxyRefs } from 'vue'

import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'

export function useGroup(axiosSource: AxiosRequestSource) {
  const { status, loading, data, reload } = useWatchAxios(() => {
    return axiosSource()
  })

  const firstSeenAt = computed((): string => {
    return data.value?.firstSeenAt
  })

  const lastSeenAt = computed((): string => {
    return data.value?.lastSeenAt
  })

  const summary = computed(() => {
    return data.value?.summary ?? {}
  })

  function getMetric(name: string): number {
    return summary.value[name] ?? 0
  }

  return proxyRefs({
    status,
    loading,
    reload,

    firstSeenAt,
    lastSeenAt,
    summary,
    getMetric,
  })
}
