import { createFormatter, createShortFormatter, Formatter } from '@/util/fmt'
import {
  EChartsOption as BaseEChartsOption,
  LegendComponentOption,
  GridComponentOption,
  XAXisComponentOption,
  YAXisComponentOption,
  DatasetComponentOption,
  SeriesOption,
  TooltipComponentOption,
} from 'echarts'

import { truncateMiddle } from '@/util/string'

export interface EChartsOption extends BaseEChartsOption {
  legend: LegendComponentOption[]
  grid: GridComponentOption[]
  xAxis: XAXisComponentOption[]
  yAxis: YAXisComponentOption[]
  dataset: DatasetComponentOption[]
  series: SeriesOption[]
  tooltip: TooltipComponentOption[]
}

export function baseChartConfig(): EChartsOption {
  return {
    animation: false,
    textStyle: {
      fontFamily: '"Roboto", sans-serif',
    },

    toolbox: { show: false },
    dataZoom: [
      {
        type: 'inside',
        disabled: true,
      },
    ],

    legend: [],
    grid: [],
    xAxis: [],
    yAxis: [],
    dataset: [],
    series: [],
    tooltip: [],
  }
}

export function addChartTooltip(cfg: any, tooltipCfg: TooltipComponentOption = {}) {
  cfg.tooltip.push({
    trigger: 'axis',
    appendToBody: true,
    axisPointer: {
      type: 'cross',
      link: [{ xAxisIndex: 'all' }],
    },
    ...tooltipCfg,
  })
}

interface TooltipFormatterConfig {
  hideDate?: boolean
  highlighted?: Record<number, boolean>
}

export function createTooltipFormatter(
  formatter: string | Formatter | Record<string, string | Formatter> = '',
  conf: TooltipFormatterConfig = {},
) {
  const cache: Record<string, Formatter> = {}

  function getFormatter(name: string): Formatter {
    let v = cache[name]
    if (!v) {
      if (typeof formatter === 'object') {
        v = createFormatter(formatter[name])
      } else {
        v = createFormatter(formatter)
      }
      cache[name] = v
    }
    return v
  }

  return (params: any): string => {
    const rows = []

    for (let p of params) {
      const name = truncateMiddle(p.seriesName, 60)
      const value = p.value[p.encode.y[0]]
      const fmt = getFormatter(p.seriesName)

      let cssClass = ''
      if (conf.highlighted && conf.highlighted[p.seriesIndex]) {
        cssClass = 'highlighted'
      }

      rows.push(
        `<tr class="${cssClass}">` +
          `<td>${p.marker}</td>` +
          `<td>${name}</td>` +
          `<td>${fmt(value)}</td>` +
          `</tr>`,
      )

      if (rows.length === 20) {
        break
      }
    }

    const ss = [
      '<div class="chart-tooltip">',
      conf.hideDate ? '' : `<p>${params[0].axisValueLabel}</p>`,
      '<table>',
      '<tbody>',
      rows.join(''),
      '</tbody>',
      '</table>',
      '</div',
    ]

    return ss.join('')
  }
}

export function axisPointerFormatter(unit = '') {
  const fmt = createFormatter(unit)
  return (params: any): string => {
    return fmt(toNumber(params.value))
  }
}

export function axisLabelFormatter(unit = '') {
  const fmt = createShortFormatter(unit)
  return (value: any): string => {
    return fmt(toNumber(value))
  }
}

function toNumber(v: any): any {
  if (typeof v === 'string') {
    const n = parseInt(v, 10)
    if (!isNaN(n)) {
      return n
    }
  }
  return v
}

export type HistogramBin = [number, number] // gte, lt

export type HeatmapPoint = [number, number, number] // [xIdx, yIdx, count]

export interface HistogramData {
  count: number[]
  bins: HistogramBin[]
  p50: number
  p90: number
  p99: number
}

export function findBinIndex(bins: HistogramBin[], x: number): number {
  return bins.findIndex((bin) => {
    return x >= bin[0] && (bin[1] === 0 || x <= bin[1])
  })
}

export function findBin(bins: HistogramBin[], x: number): HistogramBin | undefined {
  return bins.find((bin) => {
    return x >= bin[0] && (bin[1] === 0 || x <= bin[1])
  })
}
