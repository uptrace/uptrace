import { proxyRefs } from 'vue'

// Composables
import { useAxios } from '@/use/axios'

export function usePinnedFacetManager() {
  const url = `/api/v1/pinned-facets`
  const { loading: pending, request } = useAxios()

  function add(attr: string) {
    return request({ method: 'POST', url, data: { attr } })
  }

  function remove(attr: string) {
    return request({ method: 'DELETE', url, data: { attr } })
  }

  return proxyRefs({ pending, add, remove })
}
