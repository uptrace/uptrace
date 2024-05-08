<template>
  <div>
    <EChart
      :loading="loading"
      :height="percentilesChart.height"
      :option="percentilesChart.option"
      :group="internalGroup"
    />
    <EChart
      :loading="loading"
      :height="rateChart.height"
      :option="rateChart.option"
      :group="internalGroup"
      :annotations="annotations"
    />
  </div>
</template>

<script lang="ts">
import colors from 'vuetify/lib/util/colors'
import { graphic } from 'echarts'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import EChart from '@/components/EChart.vue'

// Misc
import { Unit } from '@/util/fmt'
import {
  baseChartConfig,
  addChartTooltip,
  createTooltipFormatter,
  axisLabelFormatter,
  axisPointerFormatter,
  MarkPoint,
  EChartsOption,
} from '@/util/chart'
import { Annotation } from '@/org/types'

export default defineComponent({
  name: 'PercentilesChart',
  components: { EChart },

  props: {
    selDateRange: {
      type: Object as PropType<UseDateRange>,
      default: undefined,
    },
    loading: {
      type: Boolean,
      default: false,
    },
    time: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    p50: {
      type: Array as PropType<number[]>,
      default: () => [],
    },
    p90: {
      type: Array as PropType<number[]>,
      default: () => [],
    },
    p99: {
      type: Array as PropType<number[]>,
      default: () => [],
    },
    max: {
      type: Array as PropType<number[]>,
      default: () => [],
    },
    countPerMin: {
      type: Array as PropType<number[]>,
      default: () => [],
    },
    errorsPerMin: {
      type: Array as PropType<number[]>,
      default: () => [],
    },
    group: {
      type: [String, Symbol],
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
  },

  setup(props) {
    const internalGroup = computed(() => {
      if (props.group) {
        return props.group
      }
      return Symbol()
    })

    const percentilesChart = computed(() => {
      const conf = baseChartConfig()
      addChartTooltip(conf, {
        formatter: createTooltipFormatter(Unit.Nanoseconds, { hideDate: true }),
      })

      const chart = {
        height: 110,
        option: conf,
      }

      conf.xAxis.push({
        type: 'time',
        axisTick: { show: false },
        splitLine: { show: false },
        axisLabel: { show: false },
        axisPointer: { label: { show: false } },
      })

      conf.yAxis.push({
        type: 'value',
        axisLabel: {
          formatter: axisLabelFormatter(Unit.Nanoseconds),
        },
        axisPointer: {
          label: {
            formatter: axisPointerFormatter(Unit.Nanoseconds),
          },
        },
        splitNumber: 4,
      })

      conf.dataset.push({
        source: {
          time: props.time,
          p50: props.p50,
          p90: props.p90,
          p99: props.p99,
          max: props.max,
        },
      })

      const legend: string[] = []

      const items = [
        { name: 'p50', value: props.p50, color: colors.green.lighten2 },
        { name: 'p90', value: props.p90, color: colors.orange.base },
        { name: 'p99', value: props.p99, color: colors.pink.lighten2 },
        { name: 'max', value: props.max, color: colors.red.darken2 },
      ]
      items.forEach((item, index) => {
        if (!item.value.length) {
          return
        }
        legend.push(item.name)

        conf.series.push({
          z: items.length - index,
          name: item.name,
          type: 'line',
          symbol: 'none',
          itemStyle: {
            color: item.color,
          },
          areaStyle: {
            color: new graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: item.color },
              { offset: 1, color: '#ffe' },
            ]),
          },
          encode: { x: 'time', y: item.name },
        })
      })

      conf.legend.push({
        type: 'scroll',
        width: '90%',
        data: legend,
        selected: {
          p50: true,
          p90: true,
          p99: false,
          max: false,
        },
      })

      conf.grid.push({
        top: 30,
        left: 45,
        right: 30,
        height: 75,
      })

      if (props.markPoint) {
        addMarkPoint(chart.option, props.markPoint)
      }

      return chart
    })

    const rateChart = computed(() => {
      const conf = baseChartConfig()
      addChartTooltip(conf, {
        formatter: createTooltipFormatter(Unit.None, { hideDate: true }),
      })

      const chart = {
        height: 110,
        option: conf,
      }

      conf.xAxis.push({
        type: 'time',
        axisPointer: {
          label: {
            formatter: axisPointerFormatter(Unit.Time),
          },
        },
      })

      conf.yAxis.push({
        type: 'value',
        axisLabel: {
          formatter: axisLabelFormatter(),
        },
        axisPointer: {
          label: {
            formatter: axisPointerFormatter(),
          },
        },
        splitNumber: 4,
      })

      conf.dataset.push({
        source: {
          time: props.time,
          countPerMin: props.countPerMin,
          errorsPerMin: props.errorsPerMin,
        },
      })
      conf.series.push({
        datasetIndex: conf.dataset.length - 1,
        name: 'count per min',
        type: 'line',
        symbol: 'none',
        itemStyle: { color: colors.blue.lighten1 },
        areaStyle: { opacity: 0.15 },
        encode: { x: 'time', y: 'countPerMin' },
      })

      conf.series.push({
        datasetIndex: conf.dataset.length - 1,
        name: 'errors per min',
        type: 'line',
        symbol: 'none',
        itemStyle: { color: colors.red.base },
        encode: { x: 'time', y: 'errorsPerMin' },
      })

      conf.grid.push({
        top: 15,
        left: 45,
        right: 30,
        height: 70,
      })

      return chart
    })

    return {
      internalGroup,
      percentilesChart,
      rateChart,
    }
  },
})

function addMarkPoint(conf: EChartsOption, markPoint: MarkPoint) {
  conf.series.push({
    name: markPoint.name,
    type: 'scatter',
    data: [[new Date(markPoint.time), markPoint.value]],
  })
  if (conf.legend.length) {
    // @ts-expect-error
    conf.legend[0].data.push('span')
    // @ts-expect-error
    conf.legend[0].selected.span = false
  }
}
</script>

<style lang="scss" scoped></style>
