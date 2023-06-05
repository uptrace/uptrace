import { computed, watch, ref, proxyRefs } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useAxios } from '@/use/axios'
import { useWatchAxios } from '@/use/watch-axios'
import { usePager } from '@/use/pager'
import { useForceReload } from '@/use/force-reload'
import { AttrMatcher } from '@/use/attr-matcher'

// Utilities
import { MetricAlias } from '@/metrics/types'

export interface BaseMonitor {
  id: number
  projectId: number

  type: MonitorType
  name: string
  state: MonitorState

  notifyEveryoneByEmail: boolean
  params: Record<string, any>

  createdAt: string
  updatedAt: string | null

  channelIds: number[]
  alertCount?: number
}

export enum MonitorType {
  Metric = 'metric',
  Error = 'error',
}

export interface MetricMonitor extends BaseMonitor {
  type: MonitorType.Metric
  params: MetricMonitorParams
}

export interface MetricMonitorParams {
  metrics: MetricAlias[]
  query: string
  column: string
  columnUnit: string

  forDuration: number
  forDurationUnit: MonitorDuration

  minValue: number | string | null
  maxValue: number | string | null
}

export interface ErrorMonitor extends BaseMonitor {
  type: MonitorType.Error
  params: ErrorMonitorParams
}

export interface ErrorMonitorParams {
  notifyOnNewErrors: boolean
  notifyOnRecurringErrors: boolean
  matchers: AttrMatcher[]
}

export type Monitor = MetricMonitor | ErrorMonitor

export enum MonitorState {
  Active = 'active',
  Paused = 'paused',
  Firing = 'firing',
  NoData = 'no-data',
}

export enum MonitorDuration {
  Minutes = 'minutes',
  Hours = 'hours',
}

export interface StateCount {
  state: string
  count: number
}

export type EmptyMetricMonitor = Omit<MetricMonitor, 'id' | 'projectId' | 'createdAt' | 'updatedAt'>

export function emptyMetricMonitor(): EmptyMetricMonitor {
  return {
    name: '',
    state: MonitorState.Active,

    notifyEveryoneByEmail: true,

    type: MonitorType.Metric,
    params: {
      metrics: [],
      query: '',
      column: '',
      columnUnit: '',

      forDuration: 5,
      forDurationUnit: MonitorDuration.Minutes,

      minValue: null,
      maxValue: null,
    },

    channelIds: [],
  }
}

export type EmptyErrorMonitor = Omit<ErrorMonitor, 'id' | 'projectId' | 'createdAt' | 'updatedAt'>

export function createEmptyErrorMonitor(): EmptyErrorMonitor {
  return {
    name: '',
    state: MonitorState.Active,

    notifyEveryoneByEmail: true,

    type: MonitorType.Error,
    params: {
      notifyOnNewErrors: true,
      notifyOnRecurringErrors: true,
      matchers: [],
    },

    channelIds: [],
  }
}

export type UseMonitors = ReturnType<typeof useMonitors>

export function useMonitors() {
  const route = useRoute()
  const { forceReloadParams } = useForceReload()
  const pager = usePager()

  const stateFilter = ref<MonitorState | undefined>()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/projects/${projectId}/monitors`,
      params: {
        ...forceReloadParams.value,
        ...pager.axiosParams,
        state: stateFilter.value ?? null,
      },
    }
  })

  const monitors = computed((): Monitor[] => {
    return data.value?.monitors ?? []
  })

  const states = computed((): StateCount[] => {
    return data.value?.states ?? []
  })

  const count = computed(() => {
    let count = 0
    for (let state of states.value) {
      count += state.count
    }
    return count
  })

  watch(count, (count) => {
    pager.numItem = count
  })

  return proxyRefs({
    pager,

    status,
    loading,

    items: monitors,
    count,
    states,
    stateFilter,

    reload,
  })
}

export function useMonitorManager() {
  const route = useRoute()
  const { loading: pending, request } = useAxios()

  function createMetricMonitor(monitor: Partial<MetricMonitor>) {
    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/monitors/metric`

    return request({ method: 'POST', url, data: monitor }).then((resp) => {
      return resp.data.monitor as MetricMonitor
    })
  }

  function updateMetricMonitor(monitor: MetricMonitor) {
    const { id, projectId } = monitor
    const url = `/api/v1/projects/${projectId}/monitors/${id}/metric`

    return request({ method: 'PUT', url, data: monitor }).then((resp) => {
      return resp.data.monitor as MetricMonitor
    })
  }

  function createErrorMonitor(monitor: Partial<ErrorMonitor>) {
    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/monitors/error`

    return request({ method: 'POST', url, data: monitor }).then((resp) => {
      return resp.data.monitor as ErrorMonitor
    })
  }

  function updateErrorMonitor(monitor: ErrorMonitor) {
    const { id, projectId } = monitor
    const url = `/api/v1/projects/${projectId}/monitors/${id}/error`

    return request({ method: 'PUT', url, data: monitor }).then((resp) => {
      return resp.data.monitor as ErrorMonitor
    })
  }

  function activate(monitor: BaseMonitor) {
    monitor.state = MonitorState.Active
    return updateState(monitor)
  }

  function pause(monitor: BaseMonitor) {
    monitor.state = MonitorState.Paused
    return updateState(monitor)
  }

  function updateState(monitor: BaseMonitor) {
    const { id, projectId, state } = monitor
    const url = `/api/v1/projects/${projectId}/monitors/${id}/${state}`

    return request({ method: 'PUT', url, data: monitor }).then((resp) => {
      return resp.data.monitor as Monitor
    })
  }

  function del(monitor: BaseMonitor) {
    const { id, projectId } = monitor
    const url = `/api/v1/projects/${projectId}/monitors/${id}`

    return request({ method: 'DELETE', url })
  }

  return proxyRefs({
    pending,

    createMetricMonitor,
    updateMetricMonitor,

    createErrorMonitor,
    updateErrorMonitor,

    del,
    pause,
    activate,
  })
}

export function useMetricMonitor() {
  const route = useRoute()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId, monitorId } = route.value.params
    return {
      url: `/api/v1/projects/${projectId}/monitors/${monitorId}`,
    }
  })

  const monitor = computed((): MetricMonitor | undefined => {
    return data.value?.monitor
  })

  return proxyRefs({
    status,
    loading,

    data: monitor,

    reload,
  })
}

export function useErrorMonitor() {
  const route = useRoute()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId, monitorId } = route.value.params
    return {
      url: `/api/v1/projects/${projectId}/monitors/${monitorId}`,
    }
  })

  const monitor = computed((): ErrorMonitor | undefined => {
    return data.value?.monitor
  })

  return proxyRefs({
    status,
    loading,

    data: monitor,

    reload,
  })
}

export function routeForMonitor(monitor: Monitor) {
  return {
    name: monitor.type === MonitorType.Metric ? 'MonitorMetricShow' : 'MonitorErrorShow',
    params: { monitorId: monitor.id },
  }
}
