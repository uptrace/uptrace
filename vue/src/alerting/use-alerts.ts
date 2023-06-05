import { shallowRef, computed, proxyRefs, ComputedRef } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useWatchAxios, AxiosRequestSource } from '@/use/watch-axios'
import { useAxios } from '@/use/axios'
import { useOrder } from '@/use/order'
import { Facet } from '@/use/faceted-search'

// Utilities
import { Unit } from '@/util/fmt'

// Types
import { MetricAlias } from '@/metrics/types'

interface BaseAlert {
  id: number
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

  const { status, loading, data, reload } = useWatchAxios(
    () => {
      const projectId = route.value.params.projectId
      const req = {
        url: `/api/v1/projects/${projectId}/alerts`,
        params: {
          ...order.axiosParams,
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

  return proxyRefs({
    status,
    loading,
    reload,

    order,

    items: alerts,
    facets,
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
    const alertIds = alerts.map((alert) => alert.id)

    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/alerts/${state}`
    return request({ method: 'PUT', url, data: { alertIds } })
  }

  function closeAll() {
    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/alerts/close-all`
    return request({ method: 'PUT', url })
  }

  function del(alerts: Alert[]) {
    const alertIds = alerts.map((alert) => alert.id)

    const { projectId } = route.value.params
    const url = `/api/v1/projects/${projectId}/alerts`
    return request({ method: 'DELETE', url, data: { alertIds } })
  }

  return proxyRefs({
    pending,

    toggle,
    close,
    open,
    closeAll,
    delete: del,
  })
}

export type UseAlertSelection = ReturnType<typeof useAlertSelection>

export function useAlertSelection(
  alerts: ComputedRef<Alert[]>,
  alertsOnPage: ComputedRef<Alert[]>,
) {
  const selectedAlertIds = shallowRef<number[]>([])

  const activeAlerts = computed((): Alert[] => {
    return alerts.value.filter((alert: Alert) => {
      return selectedAlertIds.value.includes(alert.id)
    })
  })

  const selectedAlertsOnPage = computed((): Alert[] => {
    return alertsOnPage.value.filter((alert: Alert) => selectedAlertIds.value.includes(alert.id))
  })

  const isFullPageSelected = computed(() => {
    return alertsOnPage.value.every((alert) => selectedAlertIds.value.includes(alert.id))
  })

  const isAllSelected = computed((): boolean => {
    return alerts.value.every((alert) => selectedAlertIds.value.includes(alert.id))
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

  function toggle(alert: Alert) {
    const index = selectedAlertIds.value.findIndex((alertId) => alertId === alert.id)
    if (index >= 0) {
      selectedAlertIds.value.splice(index, 1)
    } else {
      selectedAlertIds.value.push(alert.id)
    }
    selectedAlertIds.value = selectedAlertIds.value.slice()
  }

  function toggleAll() {
    if (isAllSelected.value) {
      selectedAlertIds.value = []
      return
    }

    selectedAlertIds.value = alerts.value.map((alert: Alert) => {
      return alert.id
    })
  }

  function togglePage() {
    if (!selectedAlertsOnPage.value.length) {
      selectedAlertIds.value = [
        ...selectedAlertIds.value,
        ...alertsOnPage.value.map((alert) => alert.id),
      ]
      return
    }

    selectedAlertsOnPage.value.map((alert) => {
      const index = selectedAlertIds.value.findIndex((alertId) => alertId === alert.id)
      if (index >= 0) {
        selectedAlertIds.value.splice(index, 1)
      }
    })
    selectedAlertIds.value = selectedAlertIds.value.slice()
  }

  function reset() {
    selectedAlertIds.value = []
  }

  return proxyRefs({
    alerts: activeAlerts,
    alertsOnPage: selectedAlertsOnPage,
    isFullPageSelected,
    isAllSelected,

    allAlerts: alerts,
    openAlerts,
    closedAlerts,
    alertIds: selectedAlertIds,

    toggle,
    togglePage,
    toggleAll,
    reset,
  })
}
