import { computed, shallowReactive, proxyRefs } from 'vue'
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'
import { BackendQueryInfo } from '@/use/uql'
import { ColumnInfo } from '@/tracing/use-explore-spans'

export type { ColumnInfo }

export interface TimeseriesGroup extends Record<string, any> {
  _id: string
  _name: string
  _query: string

  _selected: boolean
  _hovered: boolean
  _color: string
}

export function useTimeseries(reqSource: AxiosRequestSource) {
  const { status, loading, data, error, errorCode, reload } = useWatchAxios(() => {
    return reqSource()
  })

  const groups = computed((): TimeseriesGroup[] => {
    const groups: TimeseriesGroup[] = data.value?.groups ?? []

    return groups.map((group) => {
      const data: Record<string, number> = {}

      for (let key in group) {
        const value = group[key]
        if (Array.isArray(value)) {
          data['_avg_' + key] = avg(value)
        }
      }

      group = {
        ...group,
        ...data,
        _selected: true,
        _hovered: false,
        _color: '',
      }

      return shallowReactive(group)
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

  const queryInfo = computed((): BackendQueryInfo | undefined => {
    return data.value?.query
  })

  const columns = computed((): ColumnInfo[] => {
    const columns = data.value?.columns ?? []
    return columns
  })

  const groupingColumns = computed(() => {
    return columns.value.filter((col) => col.isGroup)
  })

  const metricColumns = computed(() => {
    return columns.value.filter((col) => !col.isGroup && col.isNum)
  })

  return proxyRefs({
    status,
    loading,
    error,
    errorCode,
    reload,

    groups,
    time,
    emptyValue,

    queryInfo,
    columns,
    groupingColumns,
    metricColumns,
  })
}

function avg(ns: number[]) {
  return ns.reduce((p, c) => p + c, 0) / ns.length
}
