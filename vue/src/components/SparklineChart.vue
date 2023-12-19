<template>
  <EChart
    :width="chart.width"
    :height="chart.height"
    :option="chart.option"
    :group="group"
    no-resize
  />
</template>

<script lang="ts">
import colors from 'vuetify/lib/util/colors'
import { defineComponent, computed, PropType } from 'vue'

// Components
import EChart, { EChartProps } from '@/components/EChart.vue'

// Misc
import {
  baseChartConfig,
  addChartTooltip,
  createTooltipFormatter,
  EChartsOption,
} from '@/util/chart'

export default defineComponent({
  name: 'SparklineChart',
  components: { EChart },

  props: {
    name: {
      type: String,
      required: true,
    },
    line: {
      type: Array as PropType<number[]>,
      required: true,
    },
    time: {
      type: Array as PropType<string[]>,
      required: true,
    },
    unit: {
      type: String,
      default: undefined,
    },
    color: {
      type: String,
      default: colors.blue.lighten1,
    },
    group: {
      type: [String, Symbol],
      default: undefined,
    },
  },

  setup(props) {
    const chart = computed(() => {
      const chart: Partial<EChartProps> = {
        width: 100,
        height: 30,
      }

      const conf = baseChartConfig()
      addChartTooltip(conf, {
        axisPointer: undefined,
        formatter: createTooltipFormatter(props.unit),
      })

      conf.grid.push({
        show: true,
        top: 2,
        left: 2,
        right: 2,
        height: 27,
        borderWidth: 0,
      })

      conf.xAxis.push({
        type: 'time',
        show: true,
        axisLine: { lineStyle: { color: colors.grey.lighten2 } },
        axisTick: { show: false },
        axisLabel: { show: false },
        splitLine: { show: false },
      })

      conf.yAxis.push({
        type: 'value',
        show: false,
      })

      plotLine(conf, props.name, props.line, props.time)

      chart.option = conf
      return chart
    })

    function plotLine(conf: EChartsOption, name: string, line: number[], time: string[]) {
      conf.dataset.push({
        source: {
          time,
          [name]: line,
        },
      })

      conf.series.push({
        type: 'line',
        name: name,
        encode: { x: 'time', y: name },
        showSymbol: false,
        lineStyle: { width: 1 },
        itemStyle: { color: props.color },
        areaStyle: { opacity: 0.15 },
      })
    }

    return { chart }
  },
})
</script>

<style lang="scss" scoped></style>
