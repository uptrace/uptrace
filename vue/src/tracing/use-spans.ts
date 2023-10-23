import { computed, watch, proxyRefs } from 'vue'

// Composables
import { usePager, PagerConfig } from '@/use/pager'
import { useOrder, Order } from '@/use/order'
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'
import { BackendQueryInfo } from '@/use/uql'

// Utilities
import { Span } from '@/models/span'

interface SpansConfig {
  pager?: PagerConfig
}

export type UseSpans = ReturnType<typeof useSpans>

export function useSpans(reqSource: AxiosRequestSource, conf: SpansConfig = {}) {
  const pager = usePager(conf.pager)
  const order = useOrder()

  const { status, loading, error, data, reload } = useWatchAxios(() => {
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

  const queryInfo = computed((): BackendQueryInfo | undefined => {
    return data.value?.query
  })

  watch(data, (data) => {
    pager.numItem = data?.count ?? 0
  })

  watch(
    (): Order | undefined => data.value?.order,
    (orderValue) => {
      if (orderValue) {
        order.withPausedWatch(() => {
          order.change(orderValue)
        })
      }
    },
  )

  return proxyRefs({
    pager,
    order,

    status,
    loading,
    error,
    reload,

    items: spans,
    queryInfo,
  })
}

export function useSpan(axiosReqSource: AxiosRequestSource) {
  const { loading, data } = useWatchAxios(axiosReqSource)

  const span = computed((): Span | undefined => {
    return data.value?.span
  })

  return proxyRefs({ loading, data: span })
}
