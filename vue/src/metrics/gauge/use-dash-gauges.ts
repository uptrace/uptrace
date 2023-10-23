import { omit } from 'lodash-es'
import { computed, proxyRefs, ComputedRef } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useAxios } from '@/use/axios'
import { useWatchAxios, AxiosParamsSource } from '@/use/watch-axios'
import { BackendQueryInfo } from '@/use/uql'

// Utilities
import { fmt } from '@/util/fmt'
import { DashGauge, DashKind, MetricColumn, ColumnInfo, StyledColumnInfo } from '@/metrics/types'
import { eChart as colorScheme } from '@/util/colorscheme'

export function emptyDashGauge(dashKind: DashKind): DashGauge {
  return {
    id: '',
    projectId: 0,
    dashId: '',

    dashKind,
    name: '',
    template: '',

    metrics: [],
    query: '',
    columnMap: {},
  }
}

export function useDashGauges(paramsSource: AxiosParamsSource) {
  const route = useRoute()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId, dashId } = route.value.params
    return {
      url: `/internal/v1/metrics/${projectId}/dashboards/${dashId}/gauges`,
      params: paramsSource(),
    }
  })

  const gauges = computed(() => {
    return data.value?.gauges ?? []
  })

  return proxyRefs({
    status,
    loading,

    items: gauges,

    reload,
  })
}

//------------------------------------------------------------------------------

export type UseDashGaugeQuery = ReturnType<typeof useDashGaugeQuery>

export function useDashGaugeQuery(
  axiosParamsSource: AxiosParamsSource,
  columnMap: ComputedRef<Record<string, MetricColumn>>,
) {
  const route = useRoute()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/internal/v1/metrics/${projectId}/gauge`,
      params: axiosParamsSource(),
    }
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

  const values = computed((): Record<string, any> => {
    return data.value?.values ?? {}
  })

  const query = computed((): BackendQueryInfo | undefined => {
    return data.value?.query
  })

  return proxyRefs({
    status,
    loading,

    query,
    values,
    columns: styledColumns,

    reload,
  })
}

//------------------------------------------------------------------------------

export function useDashGaugeManager() {
  const route = useRoute()
  const { loading: pending, request } = useAxios()

  function create(gauge: Partial<DashGauge>) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/gauges`

    return request({ method: 'POST', url, data: gauge }).then((resp) => {
      return resp.data.gauge as DashGauge
    })
  }

  function update(gauge: DashGauge) {
    const { id, projectId, dashId } = gauge
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/gauges/${id}`

    return request({ method: 'PUT', url, data: gauge }).then((resp) => {
      return resp.data.gauge as DashGauge
    })
  }

  function save(gauge: DashGauge) {
    if (gauge.id) {
      return update(gauge)
    }
    return create(omit(gauge, 'id'))
  }

  function del(gauge: DashGauge) {
    const { id, projectId, dashId } = gauge
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/gauges/${id}`

    return request({ method: 'DELETE', url, data: gauge }).then((resp) => {
      return resp.data.gauge as DashGauge
    })
  }

  function updateOrder(gauges: DashGauge[]) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/gauges`

    const data = gauges.map((gauge, index) => {
      return { id: gauge.id, index }
    })

    return request({ method: 'PUT', url, data })
  }

  return proxyRefs({ pending, create, update, save, del, updateOrder })
}

export function formatGauge(
  values: Record<string, number>,
  columns: StyledColumnInfo[],
  template: string,
  noData = '-',
): string {
  if (!columns.length) {
    return noData
  }

  if (template) {
    for (let col of columns) {
      const val = values[col.name]
      if (val === undefined) {
        template = template.replaceAll(varName(col.name), '-')
        continue
      }
      template = template.replaceAll(varName(col.name), fmtVar(val, col.unit))
    }
    return template
  }

  const col = columns[0]
  const val = values[col.name]
  if (val === undefined) {
    return '-'
  }
  return fmtVar(val, col.unit)
}

function varName(colName: string): string {
  return '${' + colName + '}'
}

function fmtVar(val: any, unit: string): string {
  if (unit.startsWith('{') && unit.endsWith('}')) {
    return fmt(val)
  }
  return fmt(val, unit)
}
