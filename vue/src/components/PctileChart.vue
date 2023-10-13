<template>
  <div>
    <EChart
      v-for="chart in charts"
      :key="chart.name"
      :annotations="annotations"
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
import { Annotation } from '@/org/use-annotations'

// Utilities
import { baseChartConfig, addChartTooltip, createTooltipFormatter } from '@/util/chart'
import { num, durationShort } from '@/util/fmt'
import { datetime } from '@/util/fmt/date'

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
    annotations: {
      type: Array as PropType<Annotation[]>,
      default: () => [],
    },
  },

  setup(props) {
    const charts = computed(() => {
      const charts: EChartProps[] = []

      if (props.data && props.data.p50) {
        charts.push(percentilesChart(props.data))
        charts.push(countAndErrorRateChart(props.data))
      } else {
        charts.push(eventRateChart(props.data))
      }

      return charts
    })

    return {
      charts,
      chartGroup: Symbol(),
    }
  },
})

function percentilesChart(data: ChartData) {
  const conf = baseChartConfig()
  addChartTooltip(conf, {
    formatter: createTooltipFormatter(durationShort),
  })

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
      formatter: durationShort,
    },
    axisPointer: {
      label: {
        formatter: (params: any) => durationShort(params.value),
      },
    },
    splitLine: { show: false },
  })

  conf.dataset.push({
    source: data as any,
  })

  const items = [
    { name: 'p50', color: colors.green.lighten2 },
    { name: 'p90', color: colors.orange.base },
    { name: 'p99', color: colors.pink.lighten2 },
    { name: 'max', color: colors.red.darken2 },
  ]
  for (let item of items) {
    if (!data[item.name]) {
      continue
    }

    conf.series.push({
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

  conf.legend.push({
    type: 'scroll',
    width: '90%',
    data: ['p50', 'p90', 'p99', 'max'],
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

  return {
    name: 'pctile',
    height: 110,
    option: conf,
  }
}

function countAndErrorRateChart(data: ChartData | undefined) {
  const conf = baseChartConfig()
  addChartTooltip(conf, {
    formatter: createTooltipFormatter((v: any) => String(v)),
  })

  conf.xAxis.push({
    type: 'time',
    axisPointer: {
      label: {
        formatter: (params: any) => datetime(params.value),
      },
    },
  })

  conf.yAxis.push({
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
    conf.dataset.push({
      source: {
        time: data.time,
        rate: data.rate,
      },
    })

    conf.series.push({
      datasetIndex: conf.dataset.length - 1,
      name: 'count per min',
      type: 'line',
      symbol: 'none',
      itemStyle: { color: colors.blue.lighten1 },
      areaStyle: { opacity: 0.15 },
      encode: { x: 'time', y: 'rate' },
    })

    conf.dataset.push({
      source: {
        time: data.time,
        errorRate: data.errorRate,
      },
    })

    conf.series.push({
      datasetIndex: conf.dataset.length - 1,
      name: 'errors per min',
      type: 'line',
      symbol: 'none',
      itemStyle: { color: colors.red.base },
      encode: { x: 'time', y: 'errorRate' },
    })
  }

  conf.grid.push({
    top: 15,
    left: 45,
    right: 30,
    height: 70,
  })

  return {
    name: 'rate',
    height: 110,
    option: conf,
  }
}

//------------------------------------------------------------------------------

function eventRateChart(data: ChartData | undefined) {
  const conf = baseChartConfig()
  addChartTooltip(conf)

  conf.xAxis.push({
    type: 'time',
    axisPointer: {
      label: {
        formatter: (params: any) => datetime(params.value),
      },
    },
  })

  conf.yAxis.push({
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
    conf.dataset.push({
      source: {
        time: data.time,
        rate: data.rate,
      },
    })

    conf.series.push({
      name: 'events per min',
      type: 'bar',
      itemStyle: { color: colors.blue.darken1 },
      encode: { x: 'time', y: 'rate' },
    })
  }

  conf.grid.push({
    top: 15,
    left: 45,
    right: 30,
    height: 120,
  })

  return {
    name: 'rate',
    height: 160,
    option: conf,
  }
}
</script>

<style lang="scss" scoped></style>
