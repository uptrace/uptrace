// Misc
import { MetricAlias } from '@/metrics/types'
import { AttrMatcher } from '@/use/attr-matcher'

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

  checkNumPoint: number
  timeOffset: number

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

export function emptyMetricMonitor(): MetricMonitor {
  return {
    id: 0,
    projectId: 0,

    name: '',
    state: MonitorState.Active,

    notifyEveryoneByEmail: true,

    type: MonitorType.Metric,
    params: {
      metrics: [],
      query: '',
      column: '',
      columnUnit: '',

      checkNumPoint: 5,
      timeOffset: 0,

      minValue: null,
      maxValue: null,
    },

    channelIds: [],

    createdAt: '',
    updatedAt: '',
  }
}
