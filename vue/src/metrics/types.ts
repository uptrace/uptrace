export interface Metric {
  id: string
  projectId: number

  name: string
  description: string
  unit: string
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
  Gauge = 'gauge',
  Additive = 'additive',
  Counter = 'counter',
  Histogram = 'histogram',
}

export interface MetricColumn {
  unit: string
}

export interface Timeseries {
  metric: string
  name: string
  unit: string
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
  unit: string
  isGroup?: boolean
}
