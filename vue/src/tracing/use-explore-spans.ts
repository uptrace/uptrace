import { format } from 'sql-formatter'
import { shallowRef, computed, watch, proxyRefs } from 'vue'
import { refDebounced } from '@vueuse/core'

// Composables
import { useRoute } from '@/use/router'
import { useOrder, Order } from '@/use/order'
import { useWatchAxios, AxiosParamsSource } from '@/use/watch-axios'
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

export function useGroups(axiosParamsSource: AxiosParamsSource, conf: ExploreConfig = {}) {
  const searchInput = shallowRef('')
  const debouncedSearchInput = refDebounced(searchInput, 600)
  const route = useRoute()
  const hasMore = shallowRef(false)

  const order = useOrder(
    conf.order ?? {
      column: AttrKey.spanCountPerMin,
      desc: true,
    },
  )

  const axiosParams = computed(() => {
    return axiosParamsSource()
  })

  const { status, loading, data, error, errorCode, errorMessage } = useWatchAxios(() => {
    if (!axiosParams.value) {
      return axiosParams.value
    }

    const params: Record<string, any> = {
      ...axiosParams.value,
      ...order.axiosParams,
    }

    if (debouncedSearchInput.value) {
      params.query = `${params.query} | where ${AttrKey.displayName} contains '${debouncedSearchInput.value}'`
    }

    const { projectId } = route.value.params
    return {
      url: `/api/v1/tracing/${projectId}/groups`,
      params,
    }
  })

  const lastAxiosParams = computed(() => {
    if (loading.value) {
      return { _: undefined }
    }
    return axiosParams.value
  })

  const groups = computed((): Group[] => {
    const groups: Group[] = data.value?.groups ?? []
    return groups
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

  const backendQuery = computed((): string => {
    const query = error.value?.response?.data?.query ?? ''
    if (!query) {
      return query
    }

    try {
      return format(query)
    } catch (err) {
      return query
    }
  })

  watch(
    () => data.value?.hasMore ?? false,
    (hasMoreValue) => {
      hasMore.value = hasMoreValue
    },
    { immediate: true },
  )

  watch(
    hasMore,
    (hasMore) => {
      order.ignoreAxiosParamsEnabled = !hasMore
    },
    { immediate: true },
  )

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
    status,
    loading,

    order,
    axiosParams: lastAxiosParams,

    searchInput,
    items: groups,
    hasMore,

    errorCode,
    errorMessage,
    backendQuery,

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
