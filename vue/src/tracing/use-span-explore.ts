import { format } from 'sql-formatter'
import { orderBy } from 'lodash-es'
import { computed, watch, proxyRefs } from 'vue'

// Composables
import { usePager, PagerConfig } from '@/use/pager'
import { useOrder, Order } from '@/use/order'
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'
import { QueryPart } from '@/use/uql'

// Utilities
import { AttrKey } from '@/models/otel'

export interface ColumnInfo {
  name: string
  isNum: boolean
  isGroup: boolean
}

export interface ExploreItem extends Record<string, any> {}

export type UseSpanExplore = ReturnType<typeof useSpanExplore>

interface SpanExploreConfig {
  pager?: PagerConfig
  order?: Order
}

export function useSpanExplore(reqSource: AxiosRequestSource, cfg: SpanExploreConfig = {}) {
  const pager = usePager(cfg.pager ?? { perPage: 10 })
  const order = useOrder(
    cfg.order ?? {
      column: AttrKey.spanCountPerMin,
      desc: true,
      syncQuery: true,
    },
  )

  const { loading, error, data } = useWatchAxios(
    () => {
      return reqSource()
    },
    { ignoreErrors: true },
  )

  const items = computed((): ExploreItem[] => {
    return data.value?.groups ?? []
  })

  const sortedItems = computed((): ExploreItem[] => {
    if (!order.column) {
      return items.value
    }

    const isDate = isDateField(order.column)
    return orderBy(
      items.value,
      (item: ExploreItem) => {
        const val = item[order.column!]
        return isDate ? new Date(val) : val
      },
      order.desc ? 'desc' : 'asc',
    )
  })

  const pageItems = computed((): ExploreItem[] => {
    const pageItems = sortedItems.value.slice(pager.pos.start, pager.pos.end)
    return pageItems
  })

  const queryParts = computed((): QueryPart[] => {
    return data.value?.queryParts
  })

  const columns = computed((): ColumnInfo[] => {
    let columns: ColumnInfo[] = data.value?.columns ?? []
    return columns
  })

  const groupColumns = computed((): ColumnInfo[] => {
    return columns.value.filter((col) => col.isGroup)
  })

  const plotColumns = computed((): ColumnInfo[] => {
    return columns.value.filter((col) => col.isNum)
  })

  const errorMessage = computed(() => {
    return error.value?.response?.data?.message ?? ''
  })

  const errorCode = computed(() => {
    return error.value?.response?.data?.code ?? ''
  })

  const query = computed((): string => {
    return format(error.value?.response?.data?.query ?? '')
  })

  watch(
    items,
    (items) => {
      pager.numItem = items.length
    },
    { immediate: true, flush: 'pre' },
  )

  return proxyRefs({
    pager,
    order,

    loading,

    items: sortedItems,
    pageItems,

    queryParts,
    columns,
    groupColumns,
    plotColumns,

    error,
    errorCode,
    errorMessage,
    query,
  })
}

function isDateField(s: string): boolean {
  return s === AttrKey.spanTime || hasField(s, 'time') || hasField(s, 'date')
}

function hasField(s: string, field: string): boolean {
  return s.endsWith('.' + field) || s.endsWith('_' + field)
}
