import { computed, watch, proxyRefs } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useOrder, Order } from '@/use/order'
import { useWatchAxios, AxiosRequestSource, AxiosParamsSource } from '@/use/watch-axios'
import { BackendQueryInfo } from '@/use/uql'

// Utilities
import { AttrKey } from '@/models/otel'

export interface ColumnInfo {
  name: string
  unit: string
  isNum: boolean
  isGroup: boolean
}

export interface Group extends Record<string, any> {
  _id: string
  _name: string
  _query: string
}

export type UseGroups = ReturnType<typeof useGroups>

interface ExploreConfig {
  order?: Order
}

export function useGroups(reqSource: AxiosRequestSource, conf: ExploreConfig = {}) {
  const order = useOrder(
    conf.order ?? {
      column: AttrKey.spanCountPerMin,
      desc: true,
    },
  )

  const { status, loading, data } = useWatchAxios(
    () => {
      return reqSource()
    },
    { ignoreErrors: true },
  )

  const groups = computed((): Group[] => {
    const groups: Group[] = data.value?.groups ?? []
    return groups
  })

  const hasMore = computed(() => {
    return data.value?.hasMore ?? false
  })

  const queryInfo = computed((): BackendQueryInfo | undefined => {
    return data.value?.query
  })

  const columns = computed((): ColumnInfo[] => {
    return data.value?.columns ?? []
  })

  const plottableColumns = computed((): ColumnInfo[] => {
    return columns.value.filter((col) => !col.isGroup && col.isNum)
  })

  watch(hasMore, (hasMore) => {
    order.ignoreAxiosParamsEnabled = !hasMore
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
    status,
    loading,

    order,
    items: groups,
    hasMore,

    queryInfo,
    columns,
    plottableColumns,
  })
}

export function useGroupTimeseries(paramsSource: AxiosParamsSource) {
  const route = useRoute()

  const { data } = useWatchAxios(
    () => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/tracing/${projectId}/group-stats`,
        params: paramsSource(),
        cache: true,
      }
    },
    { debounce: 500 },
  )

  const metrics = computed((): Record<string, number[]> => {
    return data.value ?? {}
  })

  const time = computed((): string[] => {
    return data.value?._time ?? []
  })

  return proxyRefs({ data: metrics, time })
}
