import { set, del } from 'vue'
import { eChart as colorScheme } from '@/util/colorscheme'

export interface Dashboard {
  id: number
  projectId: number
  templateId: string

  name: string
  pinned: boolean

  gridQuery: string

  tableMetrics: MetricAlias[]
  tableQuery: string
  tableGrouping: string[]
  tableColumnMap: Record<string, MetricColumn>
}

interface BaseGridColumn {
  id: number
  projectId: number
  dashId: number

  name: string
  description: string

  width: number
  height: number
  xAxis: number
  yAxis: number

  gridQueryTemplate: string

  type: GridColumnType
  params: unknown
}

export enum GridColumnType {
  Chart = 'chart',
  Table = 'table',
  Heatmap = 'heatmap',
}

export interface ChartGridColumn extends BaseGridColumn {
  type: GridColumnType.Chart
  params: ChartColumnParams
}

export interface ChartColumnParams {
  chartKind: ChartKind
  metrics: MetricAlias[]
  query: string
  columnMap: Record<string, MetricColumn>
  timeseriesMap: Record<string, TimeseriesStyle>
  legend: ChartLegend
}

export enum ChartKind {
  Line = 'line',
  Area = 'area',
  Bar = 'bar',
  StackedArea = 'stacked-area',
  StackedBar = 'stacked-bar',
}

export interface TimeseriesStyle {
  color: string
  opacity: number
  lineWidth: number
  symbol: string
  symbolSize: number
}

export interface ChartLegend {
  type: LegendType
  placement: LegendPlacement
  values: LegendValue[]
  maxLength?: number
}

export enum LegendType {
  None = 'none',
  List = 'list',
  Table = 'table',
}

export enum LegendPlacement {
  Bottom = 'bottom',
  Right = 'right',
}

export enum LegendValue {
  Avg = 'avg',
  Min = 'min',
  Max = 'max',
  Last = 'last',
}

export interface TableGridColumn extends BaseGridColumn {
  type: GridColumnType.Table
  params: TableColumnParams
}

export interface TableColumnParams {
  metrics: MetricAlias[]
  query: string
  columnMap: Record<string, MetricColumn>
}

export interface HeatmapGridColumn extends BaseGridColumn {
  type: GridColumnType.Heatmap
  params: HeatmapColumnParams
}

export interface HeatmapColumnParams {
  metric: string
  unit: string
  query: string
}

export type GridColumn = ChartGridColumn | TableGridColumn | HeatmapGridColumn

export enum DashKind {
  Grid = 'grid',
  Table = 'table',
}

export interface DashGauge {
  id: string
  projectId: number
  dashId: string

  dashKind: DashKind
  name: string
  template: string

  metrics: MetricAlias[]
  query: string
  columnMap: Record<string, MetricColumn>
}

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

export function emptyMetric() {
  return {
    id: '',
    projectId: 0,

    name: '',
    description: '',
    unit: '',
    instrument: Instrument.Deleted,

    createdAt: '',
    updatedAt: '',
  }
}

export interface ActiveMetric extends Metric {
  alias: string
}

export interface MetricAlias {
  name: string
  alias: string
}

export enum Instrument {
  Deleted = 'deleted',
  Gauge = 'gauge',
  Additive = 'additive',
  Counter = 'counter',
  Histogram = 'histogram',
}

export interface MetricColumn {
  unit: string
  color: string
}

export interface Timeseries {
  id: string
  metric: string
  name: string
  unit: string
  attrs: Record<string, string>
  value: number[]

  last: number
  avg: number
  min: number
  max: number
}

export interface StyledTimeseries extends Timeseries, TimeseriesStyle {}

export interface ColumnInfo {
  name: string
  isGroup?: boolean
  unit: string
}

export interface StyledColumnInfo extends ColumnInfo {
  color: string
}

export function defaultTimeseriesStyle(): TimeseriesStyle {
  return {
    color: '',
    opacity: 15,
    lineWidth: 2,
    symbol: 'none',
    symbolSize: 4,
  }
}

export function defaultChartLegend(): ChartLegend {
  return {
    type: LegendType.List,
    placement: LegendPlacement.Bottom,
    values: [LegendValue.Avg],
    maxLength: 40,
  }
}

export function updateColumnMap(colMap: Record<string, MetricColumn>, columns: ColumnInfo[]) {
  const unused = new Set(Object.keys(colMap))

  for (let col of columns) {
    if (col.isGroup) {
      continue
    }
    unused.delete(col.name)
    if (col.name in colMap) {
      continue
    }
    set(colMap, col.name, {
      unit: col.unit,
      color: '',
    })
  }

  for (let colName of unused.values()) {
    del(colMap, colName)
  }

  const colorSet = new Set(colorScheme)

  for (let colName in colMap) {
    const col = colMap[colName]
    if (col.color) {
      colorSet.delete(col.color)
    }
  }

  const colors = Array.from(colorSet)
  let index = 0
  for (let colName in colMap) {
    const col = colMap[colName]
    if (!col.color) {
      col.color = colors[index % colors.length]
      index++
    }
  }
}
