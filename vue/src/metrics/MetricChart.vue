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
      v-model="echart"
      :loading="loading"
      :height="chart.height"
      :option="chart.option"
      :group="group"
      :annotations="annotations"
    />
  </div>
</template>

<script lang="ts">
import colors from 'vuetify/lib/util/colors'
import { ECharts } from 'echarts'
import { defineComponent, shallowRef, computed, onMounted, PropType } from 'vue'

// Components
import EChart, { EChartProps } from '@/components/EChart.vue'

// Misc
import { ChartKind, StyledTimeseries } from '@/metrics/types'
import { Annotation } from '@/org/types'
import { EventBus } from '@/models/eventbus'
import { createFormatter, isCustomUnit, Unit, Formatter } from '@/util/fmt'
import {
  baseChartConfig,
  axisLabelFormatter,
  axisPointerFormatter,
  addChartTooltip,
  createTooltipFormatter,
  EChartsOption,
  MarkPoint,
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
    timeseries: {
      type: Array as PropType<StyledTimeseries[]>,
      required: true,
    },
    time: {
      type: Array as PropType<string[]>,
      required: true,
    },
    height: {
      type: Number,
      default: 200,
    },
    chartKind: {
      type: String as PropType<ChartKind>,
      default: ChartKind.Line,
    },
    group: {
      type: [String, Symbol],
      default: () => Symbol(),
    },
    minAllowedValue: {
      type: [Number, String],
      default: '',
    },
    maxAllowedValue: {
      type: [Number, String],
      default: '',
    },
    markPoint: {
      type: Object as PropType<MarkPoint>,
      default: undefined,
    },
    annotations: {
      type: Array as PropType<Annotation[]>,
      default: () => [],
    },
    eventBus: {
      type: Object as PropType<EventBus>,
      default: undefined,
    },
  },

  setup(props) {
    const echart = shallowRef<ECharts>()

    const columnFormatter = computed(() => {
      const formatter: Record<string, Formatter> = {}
      for (let ts of props.timeseries) {
        formatter[ts.name] = createFormatter(ts.unit)
      }
      if (props.markPoint) {
        formatter[props.markPoint.name] = createFormatter(props.markPoint.unit)
      }
      return formatter
    })

    const chart = computed(() => {
      const chart: Partial<EChartProps> = { height: props.height }

      if (!props.timeseries.length) {
        return chart
      }

      const conf = baseChartConfig()
      conf.toolbox = {
        feature: {
          saveAsImage: {},
        },
      }

      addChartTooltip(conf, {
        formatter: createTooltipFormatter(columnFormatter.value),
      })

      conf.xAxis.push({
        type: 'time',
        axisLabel: {
          hideOverlap: true,
        },
        axisPointer: {
          label: {
            show: false,
            formatter: axisPointerFormatter(Unit.Time),
          },
        },
      })

      props.timeseries.forEach((ts, index) => {
        plotTimeseries(conf, props.chartKind, ts, props.timeseries.length - index)
      })

      if (typeof props.minAllowedValue === 'number' || typeof props.maxAllowedValue === 'number') {
        const series = conf.series[0]
        series.markArea = {
          itemStyle: {
            color: 'green',
            opacity: 0.2,
            borderWidth: 1,
            borderType: 'dashed',
          },
          data: [
            [
              {
                name: 'Allowed values range is in green',
                xAxis: 'min',
                yAxis: props.minAllowedValue !== '' ? props.minAllowedValue : 0,
              },
              {
                xAxis: 'max',
                yAxis: props.maxAllowedValue !== '' ? props.maxAllowedValue : 'max',
              },
            ],
          ],
        }
      }

      if (props.markPoint) {
        addMarkPoint(conf, props.markPoint)
      }

      conf.grid.push({
        top: '20px',
        left: '60px',
        right: conf.yAxis.length > 1 ? '60px' : '20px',
        height: String(props.height - 50) + 'px',
      })

      chart.option = conf
      return chart
    })

    onMounted(() => {
      if (!props.eventBus) {
        return
      }

      interface HoverEvent {
        item: StyledTimeseries
        hover: boolean
      }

      props.eventBus.on('hover', (event: HoverEvent) => {
        if (!echart.value || !props.timeseries.length) {
          return
        }

        if (!event.hover) {
          echart.value.dispatchAction({ type: 'highlight', seriesIndex: 0 })
          echart.value.dispatchAction({ type: 'downplay' })
          return
        }

        const ts = event.item
        const index = props.timeseries.findIndex((el) => el.id === ts.id)
        if (index === -1) {
          echart.value.dispatchAction({ type: 'highlight', seriesIndex: 0 })
          echart.value.dispatchAction({ type: 'downplay' })
        } else {
          echart.value.dispatchAction({ type: 'highlight', seriesIndex: index })
        }
      })
    })

    function plotTimeseries(
      conf: EChartsOption,
      chartKind: ChartKind,
      ts: StyledTimeseries,
      zIndex: number,
    ) {
      conf.dataset.push({
        source: {
          time: props.time,
          value: ts.value as number[],
        },
      })

      const series: Record<string, any> = {
        yAxisIndex: yAxisIndex(conf, ts.unit),
        datasetIndex: conf.dataset.length - 1,
        name: ts.name,
        type: 'line',
        encode: { x: 'time', y: 'value' },
        symbol: ts.symbol,
        symbolSize: ts.symbolSize,
        color: ts.color,
        emphasis: { focus: 'series' },
        connectNulls: false,
      }

      switch (chartKind) {
        case ChartKind.Line:
          series.lineStyle = { width: ts.lineWidth }
          break
        case ChartKind.Area:
          series.z = zIndex
          series.lineStyle = { width: ts.lineWidth }
          series.areaStyle = { opacity: ts.opacity / 100 }
          break
        case ChartKind.Bar:
          series.type = 'bar'
          series.areaStyle = { opacity: ts.opacity / 100 }
          break
        case ChartKind.StackedArea:
          series.stack = 'all'
          series.lineStyle = { width: ts.lineWidth }
          series.areaStyle = { opacity: ts.opacity / 100 }
          break
        case ChartKind.StackedBar:
          series.type = 'bar'
          series.stack = 'all'
          series.areaStyle = { opacity: ts.opacity / 100 }
          break
      }

      conf.series.push(series)
    }

    function yAxisIndex(conf: EChartsOption, unit: string): number {
      if (isCustomUnit(unit)) {
        unit = ''
      }

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
        min: (value) => {
          const values = [0, value.min]
          if (typeof props.minAllowedValue === 'number') {
            values.push(props.minAllowedValue)
          }
          if (typeof props.maxAllowedValue === 'number') {
            values.push(props.maxAllowedValue)
          }
          return Math.min(...values)
        },
        max: (value) => {
          const values = [value.max]
          if (typeof props.minAllowedValue === 'number') {
            values.push(props.minAllowedValue)
          }
          if (typeof props.maxAllowedValue === 'number') {
            values.push(props.maxAllowedValue)
          }
          return Math.max(...values)
        },
      })

      return conf.yAxis.length - 1
    }

    return { echart, chart }
  },
})

function addMarkPoint(conf: EChartsOption, markPoint: MarkPoint) {
  conf.series.push({
    name: markPoint.name,
    type: 'scatter',
    data: [[markPoint.time, markPoint.value]],
    symbolSize: 15,
    color: colors.red.darken2,
    z: 999,
  })
}
</script>

<style lang="scss" scoped></style>
