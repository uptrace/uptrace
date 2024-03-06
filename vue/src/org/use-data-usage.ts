import { computed, proxyRefs } from 'vue'

import { useWatchAxios } from '@/use/watch-axios'

export function useDataUsage() {
  const { status, loading, data, reload } = useWatchAxios(() => {
    return {
      url: '/internal/v1/data-usage',
    }
  })

  const usage = computed(() => {
    return data.value?.usage ?? {}
  })

  const spans = computed(() => {
    return data.value?.spans ?? 0
  })

  const bytes = computed(() => {
    return data.value?.bytes ?? 0
  })

  const timeseries = computed(() => {
    return data.value?.timeseries ?? 0
  })

  const startTime = computed(() => {
    return data.value?.startTime
  })

  const endTime = computed(() => {
    return data.value?.endTime
  })

  return proxyRefs({
    status,
    loading,
    reload,

    data: usage,
    spans,
    bytes,
    timeseries,
    startTime,
    endTime,
  })
}
