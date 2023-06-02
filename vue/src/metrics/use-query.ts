import { pick, orderBy, cloneDeep } from 'lodash-es'
import { shallowRef, reactive, computed, watch, proxyRefs, Ref } from 'vue'
import { refDebounced } from '@vueuse/core'

// Composables
import { useRoute } from '@/use/router'
import { useOrder, Order } from '@/use/order'
import {
  useWatchAxios,
  AxiosRequestSource,
  AxiosParamsSource,
  AxiosParams,
} from '@/use/watch-axios'
import { BackendQueryInfo } from '@/use/uql'

// Utilities
import { escapeRe } from '@/util/string'
import { eChart as colorScheme } from '@/util/colorscheme'

// Types
import {
  defaultTimeseriesStyle,
  Timeseries,
  StyledTimeseries,
  TimeseriesStyle,
  MetricColumn,
  ColumnInfo,
  StyledColumnInfo,
} from '@/metrics/types'

export type UseTimeseries = ReturnType<typeof useTimeseries>

interface TimeseriesConfig {
  cache?: boolean
}

export function useTimeseries(axiosParamsSource: AxiosParamsSource, conf: TimeseriesConfig = {}) {
  const route = useRoute()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/metrics/${projectId}/timeseries`,
      params: axiosParamsSource(),
      cache: conf.cache ?? false,
    }
  })

  const query = computed((): BackendQueryInfo | undefined => {
    return data.value?.query
  })

  const error = computed(() => {
    const parts = query.value?.parts ?? []
    for (let part of parts) {
      if (part.error) {
        return part.error
      }
    }
    return undefined
  })

  const timeseries = computed((): Timeseries[] => {
    const timeseries: Timeseries[] = data.value?.timeseries ?? []
    return timeseries.map((ts, i) => {
      return reactive({
        ...ts,

        last: ts.value[ts.value.length - 1],
        avg: ts.value.reduce((p, c) => p + c, 0) / ts.value.length,
        min: Math.min(...ts.value),
        max: Math.max(...ts.value),
      })
    })
  })

  const time = computed((): string[] => {
    return data.value?.time ?? []
  })

  const columns = computed((): ColumnInfo[] => {
    return data.value?.columns ?? []
  })

  return proxyRefs({
    status,
    loading,
    reload,

    query,
    error,
    items: timeseries,
    time,
    columns,
  })
}

export function useStyledTimeseries(
  items: Ref<Timeseries[]>,
  columnMap: Ref<Record<string, MetricColumn>>,
  timeseriesMap: Ref<Record<string, TimeseriesStyle>>,
) {
  return computed(() => {
    const timeseries = items.value.map((ts): StyledTimeseries => {
      return {
        ...ts,
        ...defaultTimeseriesStyle(),
        ...pick(columnMap.value[ts.metric], 'unit'),
        ...timeseriesMap.value[ts.name],
      }
    })

    const colorSet = new Set(colorScheme)
    const seen = new Set()

    for (let ts of timeseries) {
      if (!ts.color) {
        continue
      }

      if (seen.has(ts.color)) {
        ts.color = ''
        continue
      }

      colorSet.delete(ts.color)
      seen.add(ts.color)
    }

    const colors = Array.from(colorSet)
    let index = 0
    for (let ts of timeseries) {
      if (!ts.color) {
        ts.color = colors[index % colors.length]
        index++
      }
    }

    return timeseries
  })
}

export interface AnalyzedTimeseries {
  name: string
  metric: string
  median: number
  minAllowedValue: number | null
  maxAllowedValue: number | null
}

export function useAnalyzedTimeseries(axiosReq: AxiosRequestSource) {
  const { status, loading, data, reload } = useWatchAxios(axiosReq)

  const timeseries = computed((): AnalyzedTimeseries[] => {
    return data.value?.timeseries ?? []
  })

  const time = computed((): string[] => {
    return data.value?.time ?? []
  })

  return proxyRefs({
    status,
    loading,

    items: timeseries,
    time,

    reload,
  })
}

//------------------------------------------------------------------------------

export interface TableItem extends Record<string, string | number> {
  _id: string
  _name: string
  _query: string
}

export type UseTableQuery = ReturnType<typeof useTableQuery>

export function useTableQuery(
  axiosParamsSource: AxiosParamsSource,
  columnMap: Ref<Record<string, MetricColumn>>,
) {
  const route = useRoute()
  const order = useOrder()

  const searchInput = shallowRef('')
  const debouncedSearchInput = refDebounced(searchInput, 1000)
  const hasMore = shallowRef(false)

  const axiosParams = computed(() => {
    return axiosParamsSource()
  })

  const { status, loading, data, reload } = useWatchAxios(() => {
    if (!axiosParams.value) {
      return axiosParams.value
    }

    const params: Record<string, any> = {
      ...axiosParams.value,
      ...order.axiosParams,
      search: debouncedSearchInput.value,
    }

    const { projectId } = route.value.params
    return {
      url: `/api/v1/metrics/${projectId}/table`,
      params,
    }
  })

  const lastAxiosParams = shallowRef<AxiosParams>()
  watch(data, () => {
    lastAxiosParams.value = cloneDeep(axiosParams.value)
  })

  const items = computed((): TableItem[] => {
    return data.value?.items ?? []
  })

  const sortedItems = computed(() => {
    const col = order.column
    if (!col) {
      return items.value
    }
    return orderBy(items.value, (item) => item[col] ?? '', order.desc ? 'desc' : 'asc')
  })

  const filteredItems = computed(() => {
    const items = sortedItems.value
    if (!searchInput.value) {
      return items
    }
    const needle = searchInput.value.toLowerCase()
    return items.filter((item) => {
      return Object.values(item._attrs).some((value) => value.toLowerCase().includes(needle))
    })
  })

  const columns = computed((): ColumnInfo[] => {
    return data.value?.columns ?? []
  })

  const styledColumns = computed((): StyledColumnInfo[] => {
    const items = columns.value.map((col) => {
      return {
        ...col,
        ...columnMap.value[col.name],
      }
    })

    const colorSet = new Set(colorScheme)

    for (let col of items) {
      if (!col.color) {
        continue
      }
      colorSet.delete(col.color)
    }

    const colors = Array.from(colorSet)
    let index = 0
    for (let col of items) {
      if (!col.color) {
        col.color = colors[index % colors.length]
        index++
      }
    }

    return items
  })

  const groupingColumns = computed((): string[] => {
    return columns.value.filter((col) => col.isGroup).map((col) => col.name)
  })

  const query = computed((): BackendQueryInfo | undefined => {
    return data.value?.query
  })

  const error = computed(() => {
    const parts = query.value?.parts ?? []
    for (let part of parts) {
      if (part.error) {
        return part.error
      }
    }
    return undefined
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
    reload,

    order,
    axiosParams: lastAxiosParams,

    items: filteredItems,
    searchInput,
    hasMore,

    query,
    error,
    columns: styledColumns,
    groupingColumns,
  })
}

//------------------------------------------------------------------------------

export type UseHeatmapQuery = ReturnType<typeof useHeatmapQuery>

export function useHeatmapQuery(axiosParamsSource: AxiosParamsSource) {
  const route = useRoute()

  const axiosParams = computed(() => {
    return axiosParamsSource()
  })

  const { status, loading, data, reload } = useWatchAxios(() => {
    if (!axiosParams.value) {
      return axiosParams.value
    }

    const { projectId } = route.value.params
    return {
      url: `/api/v1/metrics/${projectId}/heatmap`,
      params: axiosParams.value,
    }
  })

  const xAxis = computed(() => {
    return data.value?.heatmap?.xAxis ?? []
  })

  const yAxis = computed(() => {
    return data.value?.heatmap?.yAxis ?? []
  })

  const heatmapData = computed(() => {
    return data.value?.heatmap?.data ?? []
  })

  const query = computed((): BackendQueryInfo | undefined => {
    return data.value?.query
  })

  const error = computed(() => {
    const parts = query.value?.parts ?? []
    for (let part of parts) {
      if (part.error) {
        return part.error
      }
    }
    return undefined
  })

  return proxyRefs({
    status,
    loading,
    reload,

    axiosParams,
    xAxis,
    yAxis,
    data: heatmapData,

    query,
    error,
  })
}

//------------------------------------------------------------------------------

export function hasMetricAlias(query: string, alias: string): boolean {
  alias = escapeRe('$' + alias)
  return new RegExp(`${alias}([^a-z0-9]|$)`).test(query)
}
