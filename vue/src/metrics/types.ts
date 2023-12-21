import { set, del } from 'vue'
import { eChart as colorScheme } from '@/util/colorscheme'

export interface Dashboard {
  id: number
  projectId: number
  templateId: string

  name: string
  pinned: boolean

  minInterval: number
  timeOffset: number

  tableMetrics: MetricAlias[]
  tableQuery: string
  tableGrouping: string[]
  tableColumnMap: Record<string, TableColumn>

  gridQuery: string
  gridMaxWidth: number
}

export interface GridRow {
  id: number
  dashId: number

  title: string
  description: string
  expanded: boolean
  index: number

  items: GridItem[]

  createdAt: string
  updatedAt: string
}

export interface GridItemPos {
  id: number
  width: number
  height: number
  xAxis: number
  yAxis: number
}

interface BaseGridItem {
  id: number
  dashId: number
  dashKind: DashKind
  rowId: number

  title: string
  description: string

  width: number
  height: number
  xAxis: number
  yAxis: number

  type: GridItemType
  params: unknown

  createdAt: string
  updatedAt: string
}

export enum DashKind {
  Invalid = '',
  Grid = 'grid',
  Table = 'table',
}

export enum GridItemType {
  Invalid = '',
  Gauge = 'gauge',
  Chart = 'chart',
  Table = 'table',
  Heatmap = 'heatmap',
}

export function emptyBaseGridItem(): BaseGridItem {
  return {
    id: 0,
    dashId: 0,
    dashKind: DashKind.Invalid,
    rowId: 0,

    title: '',
    description: '',

    xAxis: 0,
    yAxis: 0,
    width: 0,
    height: 0,

    type: GridItemType.Invalid,
    params: undefined,

    createdAt: '',
    updatedAt: '',
  }
}

export interface ChartGridItem extends BaseGridItem {
  type: GridItemType.Chart
  params: ChartGridItemParams
}

export interface ChartGridItemParams {
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

export interface TableGridItem extends BaseGridItem {
  type: GridItemType.Table
  params: TableGridItemParams
}

export interface TableGridItemParams {
  metrics: MetricAlias[]
  query: string
  columnMap: Record<string, TableColumn>
  itemsPerPage: number
  denseTable: boolean
}

export interface HeatmapGridItem extends BaseGridItem {
  type: GridItemType.Heatmap
  params: HeatmapGridItemParams
}

export interface HeatmapGridItemParams {
  metric: string
  unit: string
  query: string
}

export interface RowGridItem extends BaseGridItem {
  type: GridItemType.Invalid
  params: RowGridItemParams
}

export interface RowGridItemParams {
  gridItems: GridItem[]
}

export interface GaugeGridItem extends BaseGridItem {
  type: GridItemType.Gauge
  params: GaugeGridItemParams
}

export interface GaugeGridItemParams {
  metrics: MetricAlias[]
  query: string
  columnMap: Record<string, GaugeColumn>

  template: string
  valueMappings: ValueMapping[]
}

export interface ValueMapping {
  op: MappingOp
  value: number
  text: string
  color: string
}

export enum MappingOp {
  Any = 'any',
  Equal = 'eq',
  Lt = 'lt',
  Lte = 'lte',
  Gt = 'gt',
  Gte = 'gte',
}

export function emptyValueMapping(): ValueMapping {
  return {
    op: MappingOp.Equal,
    value: 0,
    text: '',
    color: '',
  }
}

export type GridItem = GaugeGridItem | ChartGridItem | TableGridItem | HeatmapGridItem

export interface Metric {
  id: string
  projectId: number

  name: string
  description: string
  instrument: Instrument
  unit: string
  attrKeys: string[]

  createdAt: string
  updatedAt: string
}

export function emptyMetric() {
  return {
    id: '',
    projectId: 0,

    name: '',
    description: '',
    instrument: Instrument.Deleted,
    unit: '',
    attrKeys: [],

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
  Summary = 'summary',
}

export interface MetricColumn {
  unit: string
  color: string
}

export function emptyMetricColumn(): MetricColumn {
  return {
    unit: '',
    color: '',
  }
}

export interface GaugeColumn {
  unit: string
  aggFunc: TableAggFunc
}

export interface StyledGaugeColumn extends ColumnInfo, GaugeColumn {}

export function emptyGaugeColumn(): GaugeColumn {
  return {
    unit: '',
    aggFunc: TableAggFunc.Last,
  }
}

export interface TableColumn extends MetricColumn {
  aggFunc: TableAggFunc
  sparklineDisabled: boolean
}

export enum TableAggFunc {
  Last = 'last',
  Avg = 'avg',
  Min = 'min',
  Max = 'max',
  Sum = 'sum',
}

export const aggFuncItems = [
  { text: 'Last value', value: TableAggFunc.Last },
  { text: 'Avg value', value: TableAggFunc.Avg },
  { text: 'Min value', value: TableAggFunc.Min },
  { text: 'Max value', value: TableAggFunc.Max },
  { text: 'Sum of values', value: TableAggFunc.Sum },
]

export function emptyTableColumn(): TableColumn {
  return {
    unit: '',
    color: '',
    aggFunc: TableAggFunc.Last,
    sparklineDisabled: false,
  }
}

export interface Timeseries {
  id: string
  metric: string
  name: string
  unit: string
  attrs: Record<string, string>
  attrsHash: string
  value: (number | null)[]

  last: number | null
  avg: number | null
  min: number | null
  max: number | null
}

export interface StyledTimeseries extends Timeseries, TimeseriesStyle {}

export interface ColumnInfo {
  name: string
  unit: string
  isGroup?: boolean
  tableFunc?: string
}

export interface StyledColumnInfo extends ColumnInfo {
  color: string
}

export function defaultTimeseriesStyle(): TimeseriesStyle {
  return {
    color: '',
    opacity: 10,
    lineWidth: 1.5,
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

interface BasicColumn {
  unit: string
}

export function updateColumnMap<T extends BasicColumn>(
  colMap: Record<string, T>,
  columns: ColumnInfo[],
  empty: () => T,
) {
  const unused = new Set(Object.keys(colMap))

  for (let col of columns) {
    if (col.isGroup) {
      continue
    }

    unused.delete(col.name)
    if (col.name in colMap) {
      continue
    }

    const data: Record<string, any> = {
      ...empty(),
      unit: col.unit,
    }
    if (col.tableFunc) {
      data.aggFunc = col.tableFunc
    }

    set(colMap, col.name, data)
  }

  for (let colName of unused.values()) {
    del(colMap, colName)
  }
}

interface ColoredColumn extends BasicColumn {
  color: string
}

export function assignColors(colMap: Record<string, ColoredColumn>, columns: ColumnInfo[]) {
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

export interface TableRowData extends Record<string, any> {
  _id: string
  _name: string
  _query: string
  _hash: string
}
