<template>
  <EChart :loading="loading" :height="chart.height" :option="chart.option" />
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { Matrix } from '@/components/loki/logql'

// Components
import EChart, { EChartProps } from '@/components/EChart.vue'
import {
  baseChartConfig,
  addChartTooltip,
  createTooltipFormatter,
  EChartsOption,
} from '@/util/chart'

export default defineComponent({
  name: 'LogqlChart',
  components: { EChart },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    loading: {
      type: Boolean,
      required: true,
    },
    result: {
      type: Array as PropType<Matrix[]>,
      required: true,
    },
  },

  setup(props) {
    const chart = computed(() => {
      const chart: Partial<EChartProps> = { height: 200 }

      if (!props.result.length) {
        return chart
      }

      const cfg = baseChartConfig()

      addChartTooltip(cfg, {
        formatter: createTooltipFormatter(),
      })

      cfg.xAxis.push({
        type: 'time',
        min: props.dateRange.gte,
        max: props.dateRange.lt,
      })
      cfg.yAxis.push({
        type: 'value',
        splitLine: { show: true },
      })

      for (let line of props.result) {
        plotLine(cfg, line)
      }

      const names = cfg.series.map((series) => series.name as string)
      cfg.legend.push({
        type: 'scroll',
        padding: [5, 10],
        data: names,
      })

      cfg.grid.push({
        top: cfg.legend.length ? '50px' : '20px',
        left: '50px',
        right: '20px',
        height: cfg.legend.length ? '120px' : '150px',
      })

      chart.option = cfg
      return chart
    })

    return { chart }
  },
})

function plotLine(cfg: EChartsOption, line: Matrix) {
  const source = line.values.map((value) => [new Date(value[0] * 1000), value[1]])
  cfg.dataset.push({
    dimensions: [
      { name: 'time', type: 'time' },
      { name: 'value', type: 'number' },
    ],
    source: source,
  })

  const series: Record<string, any> = {
    datasetIndex: cfg.dataset.length - 1,
    name: lineName(line),
    type: 'line',
    encode: { x: 'time', y: 'value' },
    //symbol: 'none',
  }

  cfg.series.push(series)
}

function lineName(line: Matrix) {
  const ss = []
  for (let k in line.metric) {
    const v = line.metric[k]
    ss.push(`${k}=${v}`)
  }
  return ss.join(' ')
}
</script>

<style lang="scss" scoped></style>
