<template>
  <div>
    <EChart
      v-for="chart in charts"
      :key="chart.name"
      :loading="loading"
      :height="chart.height"
      :option="chart.option"
      :group="chartGroup"
    />
  </div>
</template>

<script lang="ts">
import * as echarts from 'echarts'
import { defineComponent, computed, PropType } from 'vue'
import colors from 'vuetify/lib/util/colors'

// Components
import EChart, { EChartProps } from '@/components/EChart.vue'

// Utilities
import { baseChartConfig, addChartTooltip, createTooltipFormatter } from '@/util/chart'
import { num, durationShort } from '@/util/fmt'
import { datetime } from '@/util/date'

interface ChartData extends Record<string, unknown> {
  count: number[]
  rate: number[]

  errorCount: number[]
  errorRate: number[]
  errorPct: number[]

  p50: number[]
  p90: number[]
  p99: number[]

  time: string[]
}

const colorSet = {
  count: colors.blue.base,
  rate: colors.blue.base,
  errorCount: colors.red.darken1,
  errorPct: colors.red.darken3,
  p50: colors.green.lighten2,
  p90: colors.amber.darken1,
  p99: colors.deepOrange.darken3,
}

export default defineComponent({
  name: 'PctileChart',
  components: { EChart },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    data: {
      type: Object as PropType<ChartData | undefined>,
      default: undefined,
    },
  },

  setup(props) {
    const charts = computed(() => {
      const charts: EChartProps[] = []

      if (props.data && props.data.p50) {
        charts.push(pctilesChart(props.data))
        charts.push(countAndErrorRateChart(props.data))
      } else {
        charts.push(rateOnlyChart(props.data))
      }

      return charts
    })

    return {
      charts,
      chartGroup: Symbol(),
    }
  },
})

function pctilesChart(data: ChartData) {
  const cfg = baseChartConfig()
  addChartTooltip(cfg, {
    formatter: createTooltipFormatter(durationShort),
  })

  cfg.xAxis.push({
    type: 'time',
    axisTick: { show: false },
    splitLine: { show: false },
    axisLabel: { show: false },
    axisPointer: { label: { show: false } },
  })

  cfg.yAxis.push({
    type: 'value',
    axisLabel: {
      formatter: durationShort,
    },
    axisPointer: {
      label: {
        formatter: (params: any) => durationShort(params.value),
      },
    },
    splitLine: { show: false },
  })

  cfg.dataset.push({
    source: data as any,
  })

  const items = [
    { name: 'p50', color: colorSet.p50 },
    { name: 'p90', color: colorSet.p90 },
    { name: 'p99', color: colorSet.p99 },
  ]
  for (let item of items) {
    if (!data[item.name]) {
      continue
    }

    cfg.series.push({
      name: item.name,
      type: 'line',
      symbol: 'none',
      itemStyle: {
        color: item.color,
      },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: item.color },
          { offset: 1, color: '#ffe' },
        ]),
      },
      encode: { x: 'time', y: item.name },
    })
  }

  cfg.legend.push({
    type: 'scroll',
    width: '90%',
    data: ['p50', 'p90', 'p99'],
    selected: {
      p50: true,
      p90: true,
      p99: false,
    },
  })

  cfg.grid.push({
    top: 30,
    left: 45,
    right: 30,
    height: 65,
  })

  return {
    name: 'pctile',
    height: 100,
    option: cfg,
  }
}

function countAndErrorRateChart(data: ChartData | undefined) {
  const cfg = baseChartConfig()
  addChartTooltip(cfg, {
    formatter: createTooltipFormatter((v: any) => String(v)),
  })

  cfg.xAxis.push({
    type: 'time',
    axisPointer: {
      label: {
        formatter: (params: any) => datetime(params.value),
      },
    },
  })

  cfg.yAxis.push({
    type: 'value',
    axisLabel: {
      formatter: num,
    },
    axisPointer: {
      label: {
        formatter: (params: any) => num(params.value),
      },
    },
    splitLine: { show: false },
  })

  if (data) {
    cfg.dataset.push({
      source: {
        time: data.time,
        rate: data.rate,
      },
    })

    cfg.series.push({
      datasetIndex: cfg.dataset.length - 1,
      name: 'count per min',
      type: 'line',
      symbol: 'none',
      itemStyle: { color: colorSet.count },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: colorSet.count },
          { offset: 1, color: '#ffe' },
        ]),
      },
      encode: { x: 'time', y: 'rate' },
    })

    cfg.dataset.push({
      source: {
        time: data.time,
        errorRate: data.errorRate,
      },
    })

    cfg.series.push({
      datasetIndex: cfg.dataset.length - 1,
      name: 'errors per min',
      type: 'line',
      symbol: 'none',
      itemStyle: { color: colorSet.errorCount },
      encode: { x: 'time', y: 'errorRate' },
    })
  }

  cfg.grid.push({
    top: 15,
    left: 45,
    right: 30,
    height: 60,
  })

  return {
    name: 'rate',
    height: 100,
    option: cfg,
  }
}

//------------------------------------------------------------------------------

function rateOnlyChart(data: ChartData | undefined) {
  const cfg = baseChartConfig()
  addChartTooltip(cfg)

  cfg.xAxis.push({
    type: 'time',
    axisPointer: {
      label: {
        formatter: (params: any) => datetime(params.value),
      },
    },
  })

  cfg.yAxis.push({
    type: 'value',
    axisLabel: {
      formatter: num,
    },
    axisPointer: {
      label: {
        formatter: (params: any) => num(params.value),
      },
    },
    splitLine: { show: false },
  })

  if (data) {
    cfg.dataset.push({
      source: {
        time: data.time,
        rate: data.rate,
      },
    })

    cfg.series.push({
      name: 'events per min',
      type: 'bar',
      itemStyle: { color: colorSet.count },
      encode: { x: 'time', y: 'rate' },
    })
  }

  cfg.grid.push({
    top: 15,
    left: 45,
    right: 30,
    height: 120,
  })

  return {
    name: 'rate',
    height: 160,
    option: cfg,
  }
}
</script>

<style lang="scss" scoped></style>
