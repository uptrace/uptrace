import { orderBy } from 'lodash-es'
import { reactive, computed, watch, proxyRefs, ComputedRef } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { usePager } from '@/use/pager'
import { useOrder, Order } from '@/use/order'
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'
import { QueryPart } from '@/use/uql'

// Utilities
import { eChart as colorSet } from '@/util/colorscheme'
import { xkey } from '@/models/otelattr'
import { escapeRe } from '@/util/string'

// Types
import { Timeseries, ColumnInfo } from '@/metrics/types'

export type UseTimeseries = ReturnType<typeof useTimeseries>

export function useTimeseries(axiosReq: AxiosRequestSource) {
  const { status, loading, data, reload } = useWatchAxios(axiosReq)

  const baseQueryParts = computed((): QueryPart[] => {
    return data.value?.baseQuery ?? []
  })

  const queryParts = computed((): QueryPart[] => {
    return data.value?.queryParts ?? []
  })

  const hasError = computed((): boolean => {
    return queryParts.value.some((part) => Boolean(part.error))
  })

  const timeseries = computed((): Timeseries[] => {
    const timeseries: Timeseries[] = data.value?.timeseries ?? []
    return timeseries.map((ts, i) => {
      return reactive({
        ...ts,

        color: colorSet[i % 9],
        last: ts.value[ts.value.length - 1],
        avg: ts.value.reduce((p, c) => p + c, 0) / ts.value.length,
        min: Math.min(...ts.value),
        max: Math.max(...ts.value),
      })
    })
  })

  return proxyRefs({
    status,
    loading,

    baseQueryParts,
    queryParts,
    hasError,
    items: timeseries,

    reload,
  })
}

//------------------------------------------------------------------------------

export interface TableItem extends Record<string, string | number> {
  [xkey.itemQuery]: string
}

export type UseTableQuery = ReturnType<typeof useTableQuery>

export function useTableQuery(
  axiosParams: ComputedRef<Record<string, any>>,
  orderConf = undefined,
) {
  const { route } = useRouter()
  const pager = usePager()
  const order = useOrder(orderConf)

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/metrics/${projectId}/table`,
      params: {
        ...axiosParams.value,
        ...order.axiosParams,
      },
    }
  })

  const items = computed((): TableItem[] => {
    return data.value?.items ?? []
  })

  const hasMore = computed(() => {
    return data.value?.hasMore ?? true
  })

  const sortedItems = computed(() => {
    const col = order.column
    if (!col) {
      return items.value
    }
    return orderBy(items.value, (item) => item[col] ?? 0, order.desc ? 'desc' : 'asc')
  })

  const pagedItems = computed((): TableItem[] => {
    const items = sortedItems.value.slice(pager.pos.start, pager.pos.end)
    return items
  })

  const columns = computed((): ColumnInfo[] => {
    return data.value?.columns ?? []
  })

  const queryParts = computed((): QueryPart[] => {
    return data.value?.queryParts ?? []
  })

  const hasError = computed((): boolean => {
    return queryParts.value.some((part) => Boolean(part.error))
  })

  watch(items, (items) => {
    pager.numItem = items.length
  })

  watch(hasMore, (hasMore) => {
    if (hasMore) {
      order.unlockAxiosParams()
    } else {
      order.lockAxiosParams()
    }
  })

  watch(
    (): Order | undefined => data.value?.order,
    (orderValue) => {
      if (orderValue) {
        order.withLockedAxiosParams(() => {
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
    items: pagedItems,

    queryParts,
    hasError,
    columns,

    reload,
  })
}

export function hasMetricAlias(query: string, alias: string): boolean {
  alias = escapeRe('$' + alias)
  return new RegExp(`${alias}([^a-z0-9]|$)`).test(query)
}

//------------------------------------------------------------------------------

export type UseGaugeQuery = ReturnType<typeof useTableQuery>

export function useGaugeQuery(axiosParamsSource: () => Record<string, any>) {
  const { route } = useRouter()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/metrics/${projectId}/gauge`,
      params: axiosParamsSource(),
    }
  })

  const values = computed((): TableItem => {
    return data.value?.values ?? {}
  })

  const columns = computed((): ColumnInfo[] => {
    return data.value?.columns ?? []
  })

  const baseQueryParts = computed((): QueryPart[] => {
    return data.value?.baseQueryParts ?? []
  })

  const queryParts = computed((): QueryPart[] => {
    return data.value?.queryParts ?? []
  })

  return proxyRefs({
    status,
    loading,

    baseQueryParts,
    queryParts,
    values,
    columns,

    reload,
  })
}
