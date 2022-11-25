import { omit } from 'lodash-es'
import { shallowRef, ref, reactive, computed, watch, proxyRefs } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useAxios } from '@/use/axios'
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'
import { useForceReload } from '@/use/force-reload'

// Types
import { MetricColumn, MetricAlias } from '@/metrics/types'

export interface Dashboard {
  id: string
  projectId: number
  templateId: number

  name: string
  baseQuery: string

  isTable: boolean
  metrics: MetricAlias[] | null
  query: string
  columnMap: Record<string, MetricColumn>
}

export interface DashGauge {
  id: string
  projectId: number
  dashId: string

  name: string
  template: string

  metrics: MetricAlias[]
  query: string
  columnMap: Record<string, MetricColumn>
}

export interface DashEntry {
  id: string
  projectId: number
  dashId: string

  name: string
  description: string
  chartType: ChartType

  metrics: MetricAlias[]
  query: string
  columnMap: Record<string, MetricColumn>
}

//------------------------------------------------------------------------------

export type UseDashboards = ReturnType<typeof useDashboards>

export function useDashboards() {
  const { route } = useRouter()
  const { forceReloadParams } = useForceReload()
  const dashboards = ref<Dashboard[]>([])

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/metrics/${projectId}/dashboards`,
      params: forceReloadParams.value,
    }
  })

  const isEmpty = computed((): boolean => {
    return status.value.hasData() && !dashboards.value.length
  })

  const active = computed((): Dashboard | undefined => {
    return dashboards.value.find((d) => d.id === route.value.params.dashId)
  })

  const tree = computed(() => {
    return buildDashTree(dashboards.value)
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
    tree,

    active,

    reload,
  })
}

interface DashCategory {
  name: string
  children: Dashboard[]
}

export type DashTree = DashCategory | Dashboard

function buildDashTree(dashboards: Dashboard[]): DashTree[] {
  if (dashboards.length <= 7) {
    return dashboards
  }

  const tree: DashTree[] = []
  const m: Record<string, DashCategory> = {}

  for (let dash of dashboards) {
    const match = dash.name.match(/^(\w+): /)
    if (!match) {
      tree.push(dash)
      continue
    }

    const categoryName = match[1]
    let category = m[categoryName]

    if (!category) {
      category = { name: categoryName, children: [] }
      m[categoryName] = category
      tree.push(category)
    }

    category.children.push({
      ...dash,
      name: dash.name.slice(match[0].length),
    })
  }

  tree.sort((a, b) => a.name.localeCompare(b.name))

  return tree
}

//------------------------------------------------------------------------------

export type UseDashboard = ReturnType<typeof useDashboard>

export function useDashboard() {
  const { route } = useRouter()
  const { forceReloadParams } = useForceReload()

  const entries = ref<DashEntry[]>([])

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId, dashId } = route.value.params
    return {
      url: `/api/v1/metrics/${projectId}/dashboards/${dashId}`,
      params: forceReloadParams.value,
    }
  })

  const dashboard = computed((): Dashboard | undefined => {
    if (data.value) {
      return reactive(data.value.dashboard)
    }
    return undefined
  })

  const isTemplate = computed((): boolean => {
    return Boolean(dashboard.value?.templateId)
  })

  const isGridFull = computed((): boolean => {
    return entries.value.length >= 20
  })

  const tableGauges = computed((): DashGauge[] => {
    return data.value?.tableGauges ?? []
  })

  const gridGauges = computed((): DashGauge[] => {
    return data.value?.gridGauges ?? []
  })

  const metrics = computed({
    get() {
      return dashboard.value?.metrics ?? []
    },
    set(v: MetricAlias[]) {
      if (dashboard.value) {
        dashboard.value.metrics = v
      }
    },
  })

  const columnMap = computed((): Record<string, MetricColumn> => {
    return dashboard.value?.columnMap ?? {}
  })

  watch(
    () => data.value?.entries ?? [],
    (entriesValue) => {
      entries.value = entriesValue.map((e: DashEntry) => reactive(e))
    },
  )

  function addGridEntry() {
    if (!dashboard.value) {
      return
    }
    entries.value.push(
      reactive(
        newEmptyDashEntry({
          dashId: dashboard.value.id,
        }),
      ),
    )
  }

  return proxyRefs({
    status,
    loading,
    axiosData: data,

    data: dashboard,
    active: dashboard,
    isTemplate,
    isGridFull,
    tableGauges,
    gridGauges,
    entries: entries,

    metrics,
    columnMap,

    addGridEntry,
    reload,
  })
}

//------------------------------------------------------------------------------

export function useDashManager() {
  const { route } = useRouter()
  const { loading: pending, request } = useAxios()

  function create(dash: Partial<Dashboard>) {
    const url = `/api/v1/metrics/${dash.projectId}/dashboards`

    const data = {
      name: dash.name,
    }

    return request({ method: 'POST', url, data }).then((resp) => {
      return resp.data.dashboard as Dashboard
    })
  }

  function update(data: Partial<Dashboard>) {
    const { projectId, dashId } = route.value.params
    const url = `/api/v1/metrics/${projectId}/dashboards/${dashId}`

    return request({ method: 'PUT', url, data }).then((resp) => {
      return resp.data.dashboard as Dashboard
    })
  }

  function clone(dash: Dashboard) {
    const url = `/api/v1/metrics/${dash.projectId}/dashboards/${dash.id}`

    return request({ method: 'POST', url }).then((resp) => {
      return resp.data.dashboard as Dashboard
    })
  }

  function del(dash: Dashboard) {
    const url = `/api/v1/metrics/${dash.projectId}/dashboards/${dash.id}`

    return request({ method: 'DELETE', url }).then((resp) => {
      return resp.data.dashboard as Dashboard
    })
  }

  return proxyRefs({ pending, create, update, clone, del })
}

export function useDashAttrs(axiosReq: AxiosRequestSource) {
  const { status, loading, data, reload } = useWatchAxios(axiosReq)

  const keys = computed((): string[] => {
    const attrs = data.value?.attrs ?? []
    return attrs
  })

  return proxyRefs({
    status,
    loading,
    keys,

    reload,
  })
}

export function useDashQueryManager(dashboard: UseDashboard) {
  const savedQuery = shallowRef('')
  const man = useDashManager()

  const pending = computed(() => {
    return man.pending
  })

  const isDirty = computed(() => {
    if (!dashboard.data || dashboard.isTemplate) {
      return false
    }
    return dashboard.data.baseQuery !== savedQuery.value
  })

  watch(
    () => dashboard.axiosData,
    (data) => {
      savedQuery.value = data?.dashboard?.baseQuery ?? ''
    },
    { immediate: true },
  )

  function save() {
    if (!dashboard.data) {
      return
    }
    man.update({ baseQuery: dashboard.data.baseQuery }).then(() => {
      dashboard.reload()
    })
  }

  return proxyRefs({ pending, isDirty, save })
}

//------------------------------------------------------------------------------

export function newEmptyDashEntry(entry: Partial<DashEntry> = {}) {
  return {
    id: '',
    projectId: 0,
    dashId: '',

    name: '',
    description: '',
    chartType: ChartType.Line,

    metrics: [],
    query: '',
    columnMap: {},

    ...entry,
  }
}

export enum ChartType {
  Line = 'line',
  Area = 'area',
  Bar = 'bar',
  StackedArea = 'stacked-area',
  StackedBar = 'stacked-bar',
}

export function useDashEntryManager() {
  const { route } = useRouter()
  const { loading: pending, request } = useAxios()

  function create(entry: Partial<DashEntry>) {
    const { projectId, dashId } = route.value.params
    const url = `/api/v1/metrics/${projectId}/dashboards/${dashId}/entries`

    return request({ method: 'POST', url, data: entry }).then((resp) => {
      return resp.data.entry as DashEntry
    })
  }

  function update(entry: DashEntry) {
    const { id, projectId, dashId } = entry
    const url = `/api/v1/metrics/${projectId}/dashboards/${dashId}/entries/${id}`

    return request({ method: 'PUT', url, data: entry }).then((resp) => {
      return resp.data.entry as DashEntry
    })
  }

  function save(entry: DashEntry) {
    if (entry.id) {
      return update(entry)
    }
    return create(omit(entry, 'id'))
  }

  function del(entry: DashEntry) {
    const { id, projectId, dashId } = entry
    const url = `/api/v1/metrics/${projectId}/dashboards/${dashId}/entries/${id}`

    return request({ method: 'DELETE', url, data: entry }).then((resp) => {
      return resp.data.entry as DashEntry
    })
  }

  function updateOrder(entries: DashEntry[]) {
    const { projectId, dashId } = route.value.params
    const url = `/api/v1/metrics/${projectId}/dashboards/${dashId}/entries`

    const data = entries.map((entry, index) => {
      return { id: entry.id, weight: entries.length - index }
    })

    return request({ method: 'PUT', url, data })
  }

  return proxyRefs({ pending, create, update, save, del, updateOrder })
}
