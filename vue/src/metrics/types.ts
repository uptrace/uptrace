import { Unit } from '@/util/fmt'

export interface Metric {
  id: string
  projectId: number

  name: string
  description: string
  unit: Unit
  instrument: Instrument

  createdAt: string
  updatedAt: string
}

export interface ActiveMetric extends Metric {
  alias: string
}

export interface MetricAlias {
  name: string
  alias: string
}

export enum Instrument {
  Invalid = 'invalid',
  Gauge = 'gauge',
  Additive = 'additive',
  Counter = 'counter',
  Histogram = 'histogram',
}

export interface MetricColumn {
  unit: Unit
}

export interface Timeseries {
  metric: string
  name: string
  unit: Unit
  attrs: Record<string, string>
  value: number[]
  time: string[]

  color: string
  last: number
  avg: number
  min: number
  max: number
}

export interface ColumnInfo {
  name: string
  unit: Unit
  isGroup?: boolean
}
