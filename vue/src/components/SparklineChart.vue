<template>
  <EChart :width="chart.width" :height="chart.height" :option="chart.option" />
</template>

<script lang="ts">
import colors from 'vuetify/lib/util/colors'
import { defineComponent, computed, PropType } from 'vue'

// Components
import EChart, { EChartProps } from '@/components/EChart.vue'

// Utilities
import { unitFromName, Unit } from '@/util/fmt'
import { baseChartConfig, addChartTooltip, createTooltipFormatter } from '@/util/chart'

export default defineComponent({
  name: 'SparklineChart',
  components: { EChart },

  props: {
    name: {
      type: String,
      required: true,
    },
    unit: {
      type: String as PropType<Unit>,
      default: undefined,
    },
    line: {
      type: Array as PropType<number[]>,
      required: true,
    },
    time: {
      type: Array as PropType<string[]>,
      required: true,
    },
  },

  setup(props) {
    const unit = computed((): Unit => {
      if (props.unit !== undefined) {
        return props.unit
      }
      return unitFromName(props.name)
    })

    const chart = computed(() => {
      const chart: Partial<EChartProps> = {
        width: 100,
        height: 30,
      }

      const conf = baseChartConfig()
      addChartTooltip(conf, {
        axisPointer: undefined,
        formatter: createTooltipFormatter(unit.value),
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

    return { chart }
  },
})

function plotLine(conf: any, name: string, line: number[], time: string[]) {
  conf.dataset.push({
    source: {
      time,
      [name]: line,
    },
  })

  const color = colors.blue.base
  conf.series.push({
    type: 'line',
    name: name,
    encode: { x: 'time', y: name },
    showSymbol: false,
    lineStyle: { width: 1 },
    itemStyle: { color },
    areaStyle: { opacity: 0.15 },
  })
}
</script>

<style lang="scss" scoped></style>
