<template>
  <EChart :loading="loading" :height="chart.height" :option="chart.option" :group="group" />
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Components
import EChart, { EChartProps } from '@/components/EChart.vue'

// Misc
import { Unit } from '@/util/fmt'
import {
  baseChartConfig,
  axisLabelFormatter,
  axisPointerFormatter,
  addChartTooltip,
  createTooltipFormatter,
  EChartsOption,
} from '@/util/chart'

export default defineComponent({
  name: 'DataUsageChart',
  components: { EChart },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    time: {
      type: Array as PropType<string[]>,
      default: undefined,
    },
    value: {
      type: Array as PropType<number[]>,
      default: () => [],
    },
    name: {
      type: String,
      required: true,
    },
    unit: {
      type: String,
      default: Unit.None,
    },
    group: {
      type: String,
      default: 'usage',
    },
  },

  setup(props) {
    const chart = computed(() => {
      const chart: Partial<EChartProps> = { height: 200 }

      if (!props.time) {
        return chart
      }

      const conf = baseChartConfig()
      plotUsage(conf)

      chart.option = conf
      return chart
    })

    function plotUsage(conf: EChartsOption) {
      addChartTooltip(conf, {
        formatter: createTooltipFormatter(props.unit),
      })

      conf.xAxis.push({
        type: 'time',
        splitLine: { show: false },
        axisPointer: {
          label: {
            formatter: axisPointerFormatter(Unit.Date),
          },
        },
      })

      conf.yAxis.push({
        type: 'value',
        splitNumber: 4,
        splitLine: { show: false },
        axisLabel: {
          formatter: axisLabelFormatter(props.unit),
        },
        axisPointer: {
          label: {
            formatter: axisPointerFormatter(props.unit),
          },
        },
      })

      conf.dataset.push({
        source: {
          time: props.time!,
          value: props.value,
        },
      })

      conf.series.push({
        name: props.name,
        type: 'line',
        encode: { x: 'time', y: 'value' },
      })

      conf.grid.push({
        top: '25px',
        left: '50px',
        right: '25px',
        height: '140px',
      })
    }

    return { chart }
  },
})
</script>

<style lang="scss" scoped></style>
