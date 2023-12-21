import { pick, orderBy, cloneDeep } from 'lodash-es'
import { shallowRef, reactive, computed, watch, proxyRefs, Ref } from 'vue'
import { refDebounced } from '@vueuse/core'

// Composables
import { useRoute } from '@/use/router'
import { useOrder, Order } from '@/use/order'
import { useWatchAxios, AxiosParams, AxiosParamsSource } from '@/use/watch-axios'
import { BackendQueryInfo } from '@/use/uql'

// Misc
import { eChart as colorScheme } from '@/util/colorscheme'

// Misc
import {
  defaultTimeseriesStyle,
  Timeseries,
  StyledTimeseries,
  TimeseriesStyle,
  MetricColumn,
  ColumnInfo,
  StyledColumnInfo,
  TableRowData,
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
      url: `/internal/v1/metrics/${projectId}/timeseries`,
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

        last: lastValue(ts.value),
        avg: avgValue(ts.value),
        min: minValue(ts.value),
        max: maxValue(ts.value),
      })
    })
  })

  const time = computed((): string[] => {
    return data.value?.time ?? []
  })

  const emptyValue = computed(() => {
    const value = time.value.slice() as unknown as number[]
    value.fill(0)
    return value
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
    emptyValue,
    columns,
  })
}

function lastValue(ns: (number | null)[]): number | null {
  for (let i = ns.length - 1; i >= 0; i--) {
    const n = ns[i]
    if (n !== null) {
      return n
    }
  }
  return null
}

function avgValue(ns: (number | null)[]): number | null {
  let sum = 0
  let count = 0

  for (let n of ns) {
    if (n !== null) {
      sum += n
      count++
    }
  }

  if (count) {
    return sum / count
  }
  return 0
}

function minValue(ns: (number | null)[]): number | null {
  let min = Number.MAX_VALUE

  for (let n of ns) {
    if (n === null) {
      continue
    }
    if (n < min) {
      min = n
    }
  }

  if (min !== Number.MAX_VALUE) {
    return min
  }
  return null
}

function maxValue(ns: (number | null)[]): number | null {
  let max = Number.MIN_VALUE

  for (let n of ns) {
    if (n === null) {
      continue
    }
    if (n > max) {
      max = n
    }
  }

  if (max !== Number.MIN_VALUE) {
    return max
  }
  return null
}

//------------------------------------------------------------------------------

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

//------------------------------------------------------------------------------

export type UseTableQuery = ReturnType<typeof useTableQuery>

export function useTableQuery(
  axiosParamsSource: AxiosParamsSource,
  columnMap: Ref<Record<string, MetricColumn>>,
) {
  const route = useRoute()
  const order = useOrder()

  const searchInput = shallowRef('')
  const debouncedSearchInput = refDebounced(searchInput, 600)
  const hasMore = shallowRef(true)

  const axiosParams = computed(() => {
    return axiosParamsSource()
  })

  const { status, loading, error, data, reload } = useWatchAxios(() => {
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
      url: `/internal/v1/metrics/${projectId}/table`,
      params,
    }
  })

  const lastAxiosParams = shallowRef<AxiosParams>()
  watch(data, () => {
    // Update axios params after data is fully loaded.
    lastAxiosParams.value = cloneDeep(axiosParams.value)
  })

  const items = computed((): TableRowData[] => {
    return data.value?.items ?? []
  })

  const sortedItems = computed(() => {
    const col = order.column
    if (!col) {
      return items.value
    }
    return orderBy(items.value, (item) => item[col] ?? '', order.desc ? 'desc' : 'asc')
  })

  const columns = computed((): ColumnInfo[] => {
    return data.value?.columns ?? []
  })

  const aggColumns = computed((): StyledColumnInfo[] => {
    const items = columns.value
      .filter((col) => !col.isGroup)
      .map((col) => {
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

  const queryError = computed(() => {
    const parts = query.value?.parts ?? []
    for (let part of parts) {
      if (part.error) {
        return part.error
      }
    }
    return undefined
  })

  watch(
    () => data.value?.hasMore ?? true,
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
    reload,

    order,
    axiosParams: lastAxiosParams,

    items: sortedItems,
    searchInput,
    hasMore,

    query,
    queryError,
    columns,
    aggColumns,
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
      url: `/internal/v1/metrics/${projectId}/heatmap`,
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
