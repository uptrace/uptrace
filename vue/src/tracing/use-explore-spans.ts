import { shallowRef, computed, watch, proxyRefs } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useOrder, Order } from '@/use/order'
import { useWatchAxios, AxiosParamsSource } from '@/use/watch-axios'
import { BackendQueryInfo } from '@/use/uql'

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
  const route = useRoute()
  const hasMore = shallowRef(false)
  const order = useOrder()

  const axiosParams = computed(() => {
    return axiosParamsSource()
  })

  const { status, loading, error, data } = useWatchAxios(() => {
    if (!axiosParams.value) {
      return axiosParams.value
    }

    const params: Record<string, any> = {
      ...axiosParams.value,
      ...order.axiosParams,
    }

    const { projectId } = route.value.params
    return {
      url: `/internal/v1/tracing/${projectId}/groups`,
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

  watch(
    () => data.value?.hasMore ?? false,
    (hasMoreValue) => {
      hasMore.value = hasMoreValue
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
    error,

    order,
    axiosParams: lastAxiosParams,

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
        url: `/internal/v1/tracing/${projectId}/group-stats`,
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
