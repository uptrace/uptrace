import { computed, proxyRefs } from 'vue'

// Composables
import { useAxios } from '@/use/axios'

export interface SpanInfo {
  projectId: number
  traceId: string
  id: string
  standalone: boolean
}

export function useTraceSearch() {
  const { loading, data, request } = useAxios()

  const span = computed((): SpanInfo | undefined => {
    return data.value?.span
  })

  function find(traceId: string) {
    request({ url: `/internal/v1/traces/search`, params: { trace_id: traceId } })
  }

  return proxyRefs({ loading, span, find })
}
