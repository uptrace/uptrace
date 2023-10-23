import { omit, mergeWith } from 'lodash-es'
import { ref, reactive, computed, watch, proxyRefs } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useAxios } from '@/use/axios'
import { useWatchAxios } from '@/use/watch-axios'
import { useForceReload } from '@/use/force-reload'

// Types
import {
  defaultChartLegend,
  Dashboard,
  GridColumn,
  GridColumnType,
  MetricColumn,
  MetricAlias,
} from '@/metrics/types'

export type UseDashboards = ReturnType<typeof useDashboards>

export function useDashboards() {
  const route = useRoute()
  const { forceReloadParams } = useForceReload()
  const dashboards = ref<Dashboard[]>([])

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/internal/v1/metrics/${projectId}/dashboards`,
      params: forceReloadParams.value,
    }
  })

  const isEmpty = computed((): boolean => {
    return status.value.hasData() && !dashboards.value.length
  })

  const active = computed((): Dashboard | undefined => {
    return dashboards.value.find((d) => String(d.id) === route.value.params.dashId)
  })

  watch(
    data,
    (data) => {
      dashboards.value = (data?.dashboards ?? []).map((d: any) => reactive(d))
    },
    { immediate: true },
  )

  return proxyRefs({
    loading,
    isEmpty,
    items: dashboards,

    active,

    reload,
  })
}

//------------------------------------------------------------------------------

export type UseDashboard = ReturnType<typeof useDashboard>

export function useDashboard() {
  const route = useRoute()
  const { forceReloadParams } = useForceReload()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId, dashId } = route.value.params
    return {
      url: `/internal/v1/metrics/${projectId}/dashboards/${dashId}`,
      params: forceReloadParams.value,
    }
  })

  const dashboard = computed((): Dashboard | undefined => {
    const dash = data.value?.dashboard
    if (!dash) {
      return undefined
    }
    dash.tableMetrics ??= []
    dash.tableGrouping ??= []
    dash.tableColumnMap ??= {}
    return dash
  })

  const grid = computed((): GridColumn[] => {
    const grid = data.value?.grid ?? []

    for (let col of grid) {
      col.width = col.width || 6
      col.height = col.height || 14

      switch (col.type) {
        case GridColumnType.Chart:
          col.params.metrics ??= []
          col.params.columnMap ??= {}
          col.params.timeseriesMap ??= {}
          col.params.legend = mergeWith(
            col.params.legend,
            defaultChartLegend(),
            (objValue, srcValue) => objValue || srcValue,
          )
          break
        case GridColumnType.Table:
          col.params.metrics ??= []
          col.params.columnMap ??= {}
          break
        case GridColumnType.Heatmap:
          break
      }
    }

    return grid
  })

  const yamlUrl = computed((): string => {
    return data.value?.yamlUrl ?? ''
  })

  const isTemplate = computed((): boolean => {
    return Boolean(dashboard.value?.templateId)
  })

  const tableMetrics = computed({
    get() {
      return dashboard.value?.tableMetrics ?? []
    },
    set(v: MetricAlias[]) {
      if (dashboard.value) {
        dashboard.value.tableMetrics = v
      }
    },
  })

  const tableColumnMap = computed((): Record<string, MetricColumn> => {
    return dashboard.value?.tableColumnMap ?? {}
  })

  return proxyRefs({
    status,
    loading,
    axiosData: data,
    reload,

    data: dashboard,
    grid,
    isTemplate,
    yamlUrl,

    tableMetrics,
    tableColumnMap,
  })
}

export function useYamlDashboard() {
  const route = useRoute()
  const { forceReloadParams } = useForceReload()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId, dashId } = route.value.params
    return {
      url: `/internal/v1/metrics/${projectId}/dashboards/${dashId}/yaml`,
      params: forceReloadParams.value,
    }
  })

  const yaml = computed(() => {
    return data.value ?? ''
  })

  return proxyRefs({
    status,
    loading,
    reload,

    yaml,
  })
}

//------------------------------------------------------------------------------

export function useDashManager() {
  const route = useRoute()
  const { loading: pending, request } = useAxios()

  function create(dash: Partial<Dashboard>) {
    const url = `/internal/v1/metrics/${dash.projectId}/dashboards`

    const data = {
      name: dash.name,
    }

    return request({ method: 'POST', url, data }).then((resp) => {
      return resp.data.dashboard as Dashboard
    })
  }

  function update(data: Partial<Dashboard>) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}`

    return request({ method: 'PUT', url, data }).then((resp) => {
      return resp.data.dashboard as Dashboard
    })
  }

  function updateTable(data: Partial<Dashboard>) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/table`

    return request({ method: 'PUT', url, data }).then((resp) => {
      return resp.data.dashboard as Dashboard
    })
  }

  function updateYaml(data: string) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/yaml`

    return request({ method: 'PUT', url, data })
  }

  function clone(dash: Dashboard) {
    const url = `/internal/v1/metrics/${dash.projectId}/dashboards/${dash.id}`

    return request({ method: 'POST', url }).then((resp) => {
      return resp.data.dashboard as Dashboard
    })
  }

  function pin(dash: Dashboard) {
    const url = `/internal/v1/metrics/${dash.projectId}/dashboards/${dash.id}/pinned`

    return request({ method: 'PUT', url })
  }

  function unpin(dash: Dashboard) {
    const url = `/internal/v1/metrics/${dash.projectId}/dashboards/${dash.id}/unpinned`

    return request({ method: 'PUT', url })
  }

  function del(dash: Dashboard) {
    const url = `/internal/v1/metrics/${dash.projectId}/dashboards/${dash.id}`

    return request({ method: 'DELETE', url }).then((resp) => {
      return resp.data.dashboard as Dashboard
    })
  }

  return proxyRefs({
    pending,
    create,
    update,
    updateTable,
    updateYaml,
    clone,
    pin,
    unpin,
    delete: del,
  })
}

//------------------------------------------------------------------------------

export function useGridColumnManager() {
  const route = useRoute()
  const { loading: pending, request } = useAxios()

  function save(gridCol: GridColumn) {
    if (gridCol.id) {
      return update(gridCol)
    }
    return create(omit(gridCol, 'id'))
  }

  function create(gridCol: Omit<GridColumn, 'id'>) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/grid`

    return request({ method: 'POST', url, data: gridCol }).then((resp) => {
      return resp.data.gridCol as GridColumn
    })
  }

  function update(gridCol: GridColumn) {
    const { id, projectId, dashId } = gridCol
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/grid/${id}`

    return request({ method: 'PUT', url, data: gridCol }).then((resp) => {
      return resp.data.gridCol as GridColumn
    })
  }

  function del(gridCol: GridColumn) {
    const { id, projectId, dashId } = gridCol
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/grid/${id}`

    return request({ method: 'DELETE', url, data: gridCol }).then((resp) => {
      return resp.data.gridCol as GridColumn
    })
  }

  interface GridColumnPos {
    id: number
    width: number
    height: number
    xAxis: number
    yAxis: number
  }

  function updateOrder(grid: GridColumnPos[]) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/grid`

    return request({ method: 'PUT', url, data: grid })
  }

  return proxyRefs({ pending, create, update, save, del, updateOrder })
}
