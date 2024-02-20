import { computed, shallowRef, watch, ref, proxyRefs } from 'vue'
import { refDebounced } from '@/use/ref-debounced'

// Composables
import { useOrder } from '@/use/order'
import { useRoute, Values } from '@/use/router'
import { useAxios } from '@/use/axios'
import { useWatchAxios } from '@/use/watch-axios'
import { usePager } from '@/use/pager'
import { injectForceReload } from '@/use/force-reload'

// Misc
import { Monitor, MonitorType, MonitorState, MetricMonitor, ErrorMonitor } from '@/alerting/types'

export function createEmptyErrorMonitor(): ErrorMonitor {
  return {
    id: 0,
    projectId: 0,

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

    createdAt: '',
    updatedAt: '',
  }
}

export interface StateCount {
  state: string
  count: number
}

export type UseMonitors = ReturnType<typeof useMonitors>

export function useMonitors() {
  const route = useRoute()
  const forceReload = injectForceReload()
  const pager = usePager()
  const order = useOrder()
  const searchInput = shallowRef('')
  const debouncedSearchInput = refDebounced(searchInput, 600)

  const stateFilter = ref<MonitorState | undefined>()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    const params: Record<string, any> = {
      ...forceReload.params,
      ...pager.axiosParams(),
      state: stateFilter.value ?? null,
      ...order.axiosParams,
    }

    if (debouncedSearchInput.value) {
      params.q = debouncedSearchInput.value
    }

    return {
      url: `/internal/v1/projects/${projectId}/monitors`,
      params,
    }
  })

  const monitors = computed((): Monitor[] => {
    return data.value?.monitors ?? []
  })

  const states = computed((): StateCount[] => {
    return data.value?.states ?? []
  })

  const count = computed(() => {
    return data.value?.count ?? 0
  })

  function queryParams() {
    const queryParams: Record<string, any> = {
      q: debouncedSearchInput.value,
      ...order.queryParams(),
    }

    return queryParams
  }

  function parseQueryParams(queryParams: Values) {
    searchInput.value = queryParams.string('q')
    debouncedSearchInput.flush()

    order.parseQueryParams(queryParams)
  }

  watch(count, (count) => {
    pager.numItem = count
  })

  return proxyRefs({
    pager,

    status,
    loading,

    items: monitors,
    searchInput,
    order,
    count,
    states,
    stateFilter,

    queryParams,
    parseQueryParams,
    reload,
  })
}

export function useMonitorManager() {
  const route = useRoute()
  const { loading: pending, request } = useAxios()

  function createMetricMonitor(monitor: Partial<MetricMonitor>) {
    const { projectId } = route.value.params
    const url = `/internal/v1/projects/${projectId}/monitors/metric`

    return request({ method: 'POST', url, data: monitor }).then((resp) => {
      return resp.data.monitor as MetricMonitor
    })
  }

  function updateMetricMonitor(monitor: MetricMonitor) {
    const { id, projectId } = monitor
    const url = `/internal/v1/projects/${projectId}/monitors/${id}/metric`

    return request({ method: 'PUT', url, data: monitor }).then((resp) => {
      return resp.data.monitor as MetricMonitor
    })
  }

  function createErrorMonitor(monitor: Partial<ErrorMonitor>) {
    const { projectId } = route.value.params
    const url = `/internal/v1/projects/${projectId}/monitors/error`

    return request({ method: 'POST', url, data: monitor }).then((resp) => {
      return resp.data.monitor as ErrorMonitor
    })
  }

  function updateErrorMonitor(monitor: ErrorMonitor) {
    const { id, projectId } = monitor
    const url = `/internal/v1/projects/${projectId}/monitors/${id}/error`

    return request({ method: 'PUT', url, data: monitor }).then((resp) => {
      return resp.data.monitor as ErrorMonitor
    })
  }

  function createMonitorFromYaml(yaml: string) {
    const { projectId } = route.value.params
    const url = `/internal/v1/projects/${projectId}/monitors/yaml`

    return request({ method: 'POST', url, data: yaml }).then((resp) => {
      return resp.data.monitors as Monitor[]
    })
  }

  function activate(monitor: Monitor) {
    monitor.state = MonitorState.Active
    return updateState(monitor)
  }

  function pause(monitor: Monitor) {
    monitor.state = MonitorState.Paused
    return updateState(monitor)
  }

  function updateState(monitor: Monitor) {
    const { id, projectId, state } = monitor
    const url = `/internal/v1/projects/${projectId}/monitors/${id}/${state}`

    return request({ method: 'PUT', url, data: monitor }).then((resp) => {
      return resp.data.monitor as Monitor
    })
  }

  function del(monitor: Monitor) {
    const { id, projectId } = monitor
    const url = `/internal/v1/projects/${projectId}/monitors/${id}`

    return request({ method: 'DELETE', url })
  }

  return proxyRefs({
    pending,

    createMetricMonitor,
    updateMetricMonitor,

    createErrorMonitor,
    updateErrorMonitor,

    createMonitorFromYaml,

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
      url: `/internal/v1/projects/${projectId}/monitors/${monitorId}`,
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
      url: `/internal/v1/projects/${projectId}/monitors/${monitorId}`,
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
    params: { monitorId: String(monitor.id) },
  }
}

export function useYamlMonitor(monitorId: number) {
  const route = useRoute()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/internal/v1/projects/${projectId}/monitors/${monitorId}/yaml`,
    }
  })

  const yaml = computed(() => {
    return data.value ?? ''
  })

  return proxyRefs({
    status,
    loading,
    reload,

    data: yaml,
  })
}
