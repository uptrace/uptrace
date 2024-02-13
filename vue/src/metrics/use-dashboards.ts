import { omit, mergeWith } from 'lodash-es'
import { ref, reactive, computed, watch, proxyRefs } from 'vue'

// Composables
import { useOrder } from '@/use/order'
import { useRoute } from '@/use/router'
import { useAxios } from '@/use/axios'
import { useWatchAxios, AxiosRequestSource, AxiosParamsSource } from '@/use/watch-axios'
import { injectForceReload } from '@/use/force-reload'

// Misc
import {
  defaultChartLegend,
  Dashboard,
  GridRow,
  GridItem,
  GridItemType,
  MetricColumn,
  MetricAlias,
} from '@/metrics/types'

export type UseDashboards = ReturnType<typeof useDashboards>

export function useDashboards(axiosParamsSource: AxiosParamsSource | undefined = undefined) {
  const route = useRoute()
  const forceReload = injectForceReload()
  const order = useOrder()
  const dashboards = ref<Dashboard[]>([])

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    const axiosParams = axiosParamsSource ? axiosParamsSource() : {}

    const params: Record<string, any> = {
      ...forceReload.params,
      ...order.axiosParams,
      ...axiosParams,
    }

    return {
      url: `/internal/v1/metrics/${projectId}/dashboards`,
      params,
    }
  })

  const active = computed((): Dashboard | undefined => {
    return dashboards.value.find((d) => String(d.id) === route.value.params.dashId)
  })

  const isEmpty = computed((): boolean => {
    return status.value.hasData() && !dashboards.value.length
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
    order,

    reload,
  })
}

//------------------------------------------------------------------------------

export type UseDashboard = ReturnType<typeof useDashboard>

export function useDashboard() {
  const route = useRoute()
  const forceReload = injectForceReload()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId, dashId } = route.value.params
    return {
      url: `/internal/v1/metrics/${projectId}/dashboards/${dashId}`,
      params: forceReload.params,
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

  const tableItems = computed((): GridItem[] => {
    const tableItems = data.value?.tableItems ?? []
    for (let gridItem of tableItems) {
      fixupGridItem(gridItem)
    }
    return tableItems
  })

  const gridRows = computed((): GridRow[] => {
    const gridRows = data.value?.gridRows ?? []
    for (let gridRow of gridRows) {
      for (let gridItem of gridRow.items) {
        fixupGridItem(gridItem)
      }
    }
    return gridRows
  })

  const gridMetrics = computed((): string[] => {
    return data.value?.gridMetrics
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
    gridRows,
    gridMetrics,
    tableItems,
    isTemplate,
    yamlUrl,

    tableMetrics,
    tableColumnMap,
  })
}

export function useYamlDashboard(reqSource: AxiosRequestSource) {
  const { status, loading, data, reload } = useWatchAxios(reqSource)

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

export function useDashboardManager() {
  const route = useRoute()
  const { loading: pending, request } = useAxios()

  function create(dash: Dashboard) {
    const { projectId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards`

    return request({ method: 'POST', url, data: dash }).then((resp) => {
      return resp.data.dashboard as Dashboard
    })
  }

  function update(dash: Dashboard) {
    const { projectId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dash.id}`

    return request({ method: 'PUT', url, data: dash })
  }

  function updateTable(data: Partial<Dashboard>) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/table`

    return request({ method: 'PUT', url, data })
  }

  function updateGrid(dash: Pick<Dashboard, 'id' | 'gridQuery'>) {
    const { projectId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dash.id}/grid`

    return request({ method: 'PUT', url, data: dash })
  }

  function createYaml(data: string) {
    const { projectId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/yaml`

    return request({ method: 'POST', url, data }).then((resp) => {
      return resp.data.dashboard as Dashboard
    })
  }

  function clone(dash: Dashboard) {
    const url = `/internal/v1/metrics/${dash.projectId}/dashboards/${dash.id}`

    return request({ method: 'POST', url }).then((resp) => {
      return resp.data.dashboard as Dashboard
    })
  }

  function reset(dash: Dashboard) {
    const url = `/internal/v1/metrics/${dash.projectId}/dashboards/${dash.id}/reset`
    return request({ method: 'PUT', url })
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
    updateGrid,
    createYaml,
    clone,
    reset,
    pin,
    unpin,
    delete: del,
  })
}

//------------------------------------------------------------------------------

export function useGridItemManager() {
  const route = useRoute()
  const { loading: pending, request } = useAxios()

  function save(gridItem: GridItem) {
    if (gridItem.id) {
      return update(gridItem)
    }
    return create(omit(gridItem, 'id'))
  }

  function create(gridItem: Omit<GridItem, 'id'>) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/grid`

    return request({ method: 'POST', url, data: gridItem }).then((resp) => {
      return resp.data.gridItem as GridItem
    })
  }

  function update(gridItem: GridItem) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/grid/${gridItem.id}`

    return request({ method: 'PUT', url, data: gridItem }).then((resp) => {
      return resp.data.gridItem as GridItem
    })
  }

  function del(gridItem: GridItem) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/grid/${gridItem.id}`

    return request({ method: 'DELETE', url, data: gridItem }).then((resp) => {
      return resp.data.gridItem as GridItem
    })
  }

  return proxyRefs({ pending, create, update, save, delete: del })
}

//------------------------------------------------------------------------------

export function useGridRow(axiosReqSource: AxiosRequestSource) {
  const { status, loading, resultId, data, reload } = useWatchAxios(axiosReqSource)

  const gridRow = computed((): GridRow | undefined => {
    return data.value?.gridRow
  })

  const gridItems = computed((): GridItem[] => {
    const gridItems = data.value?.gridItems ?? []
    for (let gridItem of gridItems) {
      fixupGridItem(gridItem)
    }
    return gridItems
  })

  return proxyRefs({
    status,
    loading,
    resultId,
    reload,

    data: gridRow,
    items: gridItems,
  })
}

function fixupGridItem(gridItem: GridItem) {
  switch (gridItem.type) {
    case GridItemType.Chart:
      gridItem.params.metrics ??= []
      gridItem.params.columnMap ??= {}
      gridItem.params.timeseriesMap ??= {}
      gridItem.params.legend = mergeWith(
        gridItem.params.legend,
        defaultChartLegend(),
        (objValue, srcValue) => objValue || srcValue,
      )
      break
    case GridItemType.Table:
      gridItem.params.metrics ??= []
      gridItem.params.columnMap ??= {}
      break
    case GridItemType.Heatmap:
      break
    case GridItemType.Gauge:
      gridItem.params.metrics ??= []
      gridItem.params.columnMap ??= {}
      gridItem.params.valueMappings ??= []
      break
  }
}

export function useGridRowManager() {
  const route = useRoute()
  const { loading: pending, request } = useAxios()

  function save(gridRow: GridRow) {
    if (gridRow.id) {
      return update(gridRow)
    }
    return create(omit(gridRow, 'id'))
  }

  function create(gridRow: Partial<GridRow>) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/rows`

    return request({ method: 'POST', url, data: gridRow }).then((resp) => {
      return resp.data.gridRow as GridRow
    })
  }

  function update(gridRow: GridRow) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/rows/${gridRow.id}`

    return request({ method: 'PUT', url, data: gridRow }).then((resp) => {
      return resp.data.gridRow as GridRow
    })
  }

  function del(gridRow: GridRow) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/rows/${gridRow.id}`

    return request({ method: 'DELETE', url, data: gridRow }).then((resp) => {
      return resp.data.gridRow as GridItem
    })
  }

  function moveUp(gridRow: GridRow) {
    return move(gridRow, 'up')
  }

  function moveDown(gridRow: GridRow) {
    return move(gridRow, 'down')
  }

  function move(gridRow: GridRow, verb: string) {
    const { projectId, dashId } = route.value.params
    const url = `/internal/v1/metrics/${projectId}/dashboards/${dashId}/rows/${gridRow.id}/${verb}`
    return request({ method: 'PUT', url, data: gridRow }).then((resp) => {
      return resp.data.gridRow as GridRow
    })
  }

  return proxyRefs({ pending, create, update, save, delete: del, moveUp, moveDown })
}
