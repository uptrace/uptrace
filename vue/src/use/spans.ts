import { computed, watch, proxyRefs } from 'vue'

// Composables
import { usePager, PagerConfig } from '@/use/pager'
import { useOrder, OrderConfig } from '@/use/order'
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'
import { QueryPart } from '@/use/uql'

// Utilities
import { Span } from '@/models/span'

interface SpansConfig {
  pager?: PagerConfig
  order?: OrderConfig
}

export type UseSpans = ReturnType<typeof useSpans>

export function useSpans(reqSource: AxiosRequestSource, cfg: SpansConfig = {}) {
  const pager = usePager(cfg.pager)
  const order = useOrder(cfg.order)

  const { loading, data } = useWatchAxios(() => {
    const req = reqSource()
    if (!req) {
      return req
    }
    req.params = {
      ...req.params,
      ...order.axiosParams,
      ...pager.axiosParams(),
    }
    return req
  })

  const spans = computed((): Span[] => {
    const spans = data.value?.spans ?? []
    return spans
  })

  const queryParts = computed((): QueryPart[] => {
    return data.value?.queryParts
  })

  watch(
    data,
    (data) => {
      pager.numItem = data?.count ?? 0
    },
    { immediate: true, flush: 'sync' },
  )

  return proxyRefs({
    pager,
    order,

    loading,
    items: spans,
    queryParts,
  })
}
