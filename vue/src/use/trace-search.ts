import { computed, proxyRefs } from 'vue'

// Composables
import { useAxios } from '@/use/axios'

export interface TraceInfo {
  id: string
  projectId: number
}

export function useTraceSearch() {
  const { loading, data, request } = useAxios()

  const trace = computed((): TraceInfo | undefined => {
    return data.value?.trace
  })

  function find(traceId: string) {
    request({ url: `/api/traces/search`, params: { trace_id: traceId } })
  }

  return proxyRefs({ loading, trace, find })
}
