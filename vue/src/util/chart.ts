import { truncate } from 'lodash'
import * as echarts from 'echarts'

export function baseChartConfig(): any {
  return {
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

export function addChartTooltip(cfg: any, tooltipCfg: echarts.TooltipComponentOption = {}) {
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

export function createTooltipFormatter(fmt: (value: any) => string) {
  return (params: any): string => {
    const rows = []

    for (let p of params) {
      const name = truncate(p.seriesName, { length: 60 })
      const value = p.value[p.encode.y[0]]

      rows.push(
        `<tr>` + `<td>${p.marker}</td>` + `<td>${name}</td>` + `<td>${fmt(value)}</td>` + `</tr>`,
      )

      if (rows.length === 20) {
        break
      }
    }

    const ss = [
      '<div class="chart-tooltip">',
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
