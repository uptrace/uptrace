import { shallowRef, computed, watch, proxyRefs, ComputedRef } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'
import { useAxios } from '@/use/axios'
import { useOrder } from '@/use/order'
import { usePager } from '@/use/pager'
import { Facet } from '@/use/faceted-search'

// Utilities
import { Unit } from '@/util/fmt'

// Types
import { MetricAlias } from '@/metrics/types'

interface BaseAlert {
  id: string
  projectId: number

  name: string
  type: AlertType
  state: AlertState

  trackableId: string
  trackableModel: string

  attrs: Record<string, string>
  params: unknown

  createdAt: string
  updatedAt: string
}

export enum AlertType {
  Error = 'error',
  Metric = 'metric',
}

export enum AlertState {
  Open = 'open',
  Closed = 'closed',
}

export interface ErrorAlert extends BaseAlert {
  type: AlertType.Error
  params: {
    spanId: string
    traceId: string
    spanCount: number
  }
}

export interface MetricAlert extends BaseAlert {
  type: AlertType.Metric
  params: MetricAlertParams
}

export interface MetricAlertParams {
  firing: number
  outlier: number
  minutes: number

  monitor: {
    metrics: MetricAlias[]
    query: string
    column: string
    columnUnit: Unit
  }
}

export interface AnomalyDetector {
  median: number
  mean: number
  leftDev: number
  rightDev: number
}

export type Alert = ErrorAlert | MetricAlert

export type UseAlerts = ReturnType<typeof useAlerts>

export function useAlerts(axiosParams: ComputedRef<Record<string, any>>) {
  const route = useRoute()
  const order = useOrder({ column: 'updated_at', desc: true })
  order.syncQueryParams()
  const pager = usePager()

  const { status, loading, data, reload } = useWatchAxios(
    () => {
      const projectId = route.value.params.projectId
      const req = {
        url: `/api/v1/projects/${projectId}/alerts`,
        params: {
          ...order.axiosParams,
          ...pager.axiosParams(),
          ...axiosParams.value,
        },
      }
      return req
    },
    { debounce: 500 },
  )

  const alerts = computed((): Alert[] => {
    return data.value?.alerts ?? []
  })

  const facets = computed((): Facet[] => {
    return data.value?.facets ?? []
  })

  const resolvedCount = computed((): number => {
    return data.value?.resolvedCount ?? 0
  })

  const unresolvedCount = computed((): number => {
    return data.value?.unresolvedCount ?? 0
  })

  watch(data, (data) => {
    if (data) {
      pager.numItem = data.count
    }
  })

  return proxyRefs({
    order,
    pager,

    status,
    loading,
    items: alerts,
    facets,
    resolvedCount,
    unresolvedCount,

    reload,
  })
}

export function useAlert(reqSource: AxiosRequestSource) {
  const { status, loading, data, reload } = useWatchAxios(reqSource)

  const alert = computed((): Alert | undefined => {
    return data.value?.alert
  })

  return proxyRefs({ status, loading, data: alert, reload })
}

export function useAlertManager() {
  const route = useRoute()

  const { loading: pending, request } = useAxios()

  function toggle(alert: Alert) {
    if (alert.state === AlertState.Closed) {
      return open([alert])
    }
    return close([alert])
  }

  function close(alerts: Alert[]) {
    return toggleAlerts(alerts, AlertState.Closed)
  }

  function open(alerts: Alert[]) {
    return toggleAlerts(alerts, AlertState.Open)
  }

  function toggleAlerts(alerts: Alert[], state: AlertState) {
    const alertIds = alerts.map((alert) => {
      return alert.id
    })

    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/alerts/${state}`
    return request({ method: 'PUT', url, data: { alertIds } })
  }

  function closeAll() {
    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/alerts/close-all`
    return request({ method: 'PUT', url })
  }

  return proxyRefs({
    pending,

    toggle,
    close,
    open,
    closeAll,
  })
}

export type UseAlertSelection = ReturnType<typeof useAlertSelection>

export function useAlertSelection(alerts: ComputedRef<Alert[]>) {
  const alertIds = shallowRef<string[]>([])

  const hasOpen = computed((): boolean => {
    return alerts.value.some((alert: Alert) => alert.state !== AlertState.Closed)
  })

  const activeAlerts = computed((): Alert[] => {
    return alerts.value.filter((alert: Alert) => {
      return alertIds.value.indexOf(alert.id) !== -1
    })
  })

  const closedAlerts = computed((): Alert[] => {
    return activeAlerts.value.filter((alert: Alert) => {
      return alert.state === AlertState.Closed
    })
  })

  const openAlerts = computed((): Alert[] => {
    return activeAlerts.value.filter((alert: Alert) => {
      return alert.state !== AlertState.Closed
    })
  })

  function toggleAll() {
    if (alertIds.value.length != 0) {
      alertIds.value = []
      return
    }

    alertIds.value = alerts.value.map((alert: Alert) => {
      return alert.id
    })
  }

  function reset() {
    alertIds.value = []
  }

  return proxyRefs({
    alertIds,
    hasOpen,
    closedAlerts,
    openAlerts,
    alerts: activeAlerts,

    toggleAll,
    reset,
  })
}
