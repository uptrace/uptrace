<template>
  <div>
    <template v-if="!chart.option">
      <v-card
        v-if="resolved"
        :height="chart.height"
        flat
        class="d-flex justify-center align-center"
      >
        <div class="text-h3 grey--text text--lighten-2">NO DATA</div>
      </v-card>
      <v-skeleton-loader
        v-else
        :height="chart.height"
        type="image"
        :boilerplate="!loading"
      ></v-skeleton-loader>
    </template>

    <EChart
      v-else
      :loading="loading"
      :height="chart.height"
      :option="chart.option"
      :group="group"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { ChartType } from '@/metrics/use-dashboards'
import { MetricColumn } from '@/metrics/use-metrics'
import { Timeseries } from '@/metrics/use-query'

// Components
import EChart, { EChartProps } from '@/components/EChart.vue'

// Utilities
import { createFormatter, Unit, Formatter } from '@/util/fmt'
import {
  baseChartConfig,
  axisLabelFormatter,
  axisPointerFormatter,
  addChartTooltip,
  createTooltipFormatter,
  EChartsOption,
} from '@/util/chart'

export default defineComponent({
  name: 'MetricChart',
  components: { EChart },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    resolved: {
      type: Boolean,
      default: false,
    },
    chartType: {
      type: String as PropType<ChartType>,
      default: ChartType.Line,
    },
    columnMap: {
      type: Object as PropType<Record<string, MetricColumn>>,
      required: true,
    },
    timeseries: {
      type: Array as PropType<Timeseries[]>,
      required: true,
    },
    group: {
      type: [String, Symbol],
      default: () => Symbol(),
    },
    showLegend: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const columnFormatter = computed(() => {
      const formatter: Record<string, Formatter> = {}
      for (let ts of props.timeseries) {
        const col = props.columnMap[ts.metric]
        formatter[ts.name] = createFormatter(col?.unit ?? ts.unit)
      }
      return formatter
    })

    const chart = computed(() => {
      const chart: Partial<EChartProps> = { height: 200 }

      if (!props.timeseries.length) {
        return chart
      }

      const conf = baseChartConfig()

      addChartTooltip(conf, {
        formatter: createTooltipFormatter(columnFormatter.value),
      })

      conf.xAxis.push({
        type: 'time',
        axisPointer: {
          label: {
            show: false,
            formatter: axisPointerFormatter(Unit.Date),
          },
        },
      })

      props.timeseries.forEach((ts, index) => {
        const col = props.columnMap[ts.metric]
        plotTimeseries(
          conf,
          props.chartType,
          ts,
          col?.unit ?? ts.unit,
          props.timeseries.length - index,
        )
      })

      if (props.chartType === ChartType.Line && conf.series.length === 1) {
        const series = conf.series[0]
        if (series.type === 'line') {
          series.areaStyle = { opacity: 0.15 }
        }
      }

      if (props.showLegend && conf.series.length) {
        const names = conf.series.map((series) => series.name as string)
        conf.legend.push({
          type: 'scroll',
          padding: [5, 10],
          data: names,
        })
      }

      conf.grid.push({
        top: conf.legend.length ? '50px' : '20px',
        left: '50px',
        right: conf.yAxis.length === 2 ? '50px' : '20px',
        height: conf.legend.length ? '120px' : '150px',
      })

      chart.option = conf
      return chart
    })

    return { chart }
  },
})

function plotTimeseries(
  conf: EChartsOption,
  chartType: ChartType,
  ts: Timeseries,
  unit: Unit,
  zIndex: number,
) {
  conf.dataset.push({
    source: {
      time: ts.time,
      data: ts.value,
    },
  })

  const series: Record<string, any> = {
    yAxisIndex: yAxisIndex(conf, unit),
    datasetIndex: conf.dataset.length - 1,
    name: ts.name,
    type: 'line',
    encode: { x: 'time', y: 'data' },
    symbol: 'none',
    color: ts.color,
  }

  switch (chartType) {
    case ChartType.Area:
      series.z = zIndex
      series.areaStyle = {}
      break
    case ChartType.Bar:
      series.type = 'bar'
      break
    case ChartType.StackedArea:
      series.areaStyle = {}
      series.stack = 'all'
      series.emphasis = { focus: 'series' }
      break
    case ChartType.StackedBar:
      series.type = 'bar'
      series.stack = 'all'
      break
  }

  conf.series.push(series)
}

function yAxisIndex(conf: EChartsOption, unit: Unit): number {
  const index = conf.yAxis.findIndex((yAxis) => yAxis.id === unit)
  if (index >= 0) {
    return index
  }

  conf.yAxis.push({
    id: unit,
    type: 'value',
    axisLabel: {
      formatter: axisLabelFormatter(unit),
    },
    axisPointer: {
      label: {
        formatter: axisPointerFormatter(unit),
      },
    },
    splitLine: { show: false },
  })

  return conf.yAxis.length - 1
}
</script>

<style lang="scss" scoped></style>
