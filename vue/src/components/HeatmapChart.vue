<template>
  <div v-element-resize class="heatmap" @resize="onResize">
    <div :style="{ maxWidth: chart.width + 'px', margin: 'auto' }">
      <template v-if="!chart.option">
        <v-card
          v-if="resolved"
          :height="chart.height"
          flat
          class="d-flex justify-center align-center"
        >
          <div class="text-h3 grey--text text--lighten-2">NO DATA</div>
        </v-card>
        <v-skeleton-loader v-else type="image" :boilerplate="!loading"></v-skeleton-loader>
      </template>

      <EChart
        v-else
        v-model="echart"
        :loading="loading"
        :height="chart.height"
        :option="chart.option"
      />
    </div>
  </div>
</template>

<script lang="ts">
import * as echarts from 'echarts'
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Components
import EChart, { EChartProps } from '@/components/EChart.vue'

// Utilities
import { num } from '@/util/fmt/num'
import { datetime, toLocal } from '@/util/fmt/date'
import { createFormatter, createShortFormatter } from '@/util/fmt'
import { baseChartConfig, EChartsOption, HistogramBin, HeatmapPoint } from '@/util/chart'

export default defineComponent({
  name: 'HeatmapChart',
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
    unit: {
      type: String,
      default: '',
    },
    xAxis: {
      type: Array as PropType<string[]>,
      required: true,
    },
    yAxis: {
      type: Array as PropType<HistogramBin[]>,
      required: true,
    },
    data: {
      type: Array as PropType<HeatmapPoint[]>,
      required: true,
    },
  },

  setup(props) {
    const echart = shallowRef<echarts.ECharts>()
    const itemSize = shallowRef(10)

    const chart = computed(() => {
      const width = itemSize.value * (props.xAxis.length || 60)
      const height = itemSize.value * (props.yAxis.length || 16)
      const chart: Partial<EChartProps> = {
        width: width + 70,
        height: height + 50,
      }

      if (!props.data.length) {
        return chart
      }

      const conf = baseChartConfig()
      heatmapChart(conf)

      conf.grid.push({
        top: 15,
        left: 50,
        right: 20,
        height,
      })

      const fmt = createFormatter(props.unit)
      conf.tooltip.push({
        appendToBody: true,
        formatter(param: any) {
          const data = param.data
          const x = props.xAxis[data[0]]
          const y = props.yAxis[data[1]]
          const count = data[2]

          const ss = []
          ss.push(`${num(count)} occurrences`)
          ss.push(`${fmt(y[0])} - ${fmt(y[1])}`)
          ss.push(`${datetime(x)}`)
          const s = ss.join('<br />')
          return `<div class="chart-tooltip">${s}</div`
        },
      })

      chart.option = conf
      return chart
    })

    function heatmapChart(conf: EChartsOption) {
      const itemPadding = itemSize.value >= 14 ? 2 : 1.5
      conf.xAxis.push({
        type: 'category',
        data: props.xAxis,
        offset: itemPadding,
        axisTick: { alignWithLabel: true },
        axisLabel: {
          interval: 10,
          formatter(value: string) {
            const dt = toLocal(new Date(value))
            return echarts.time.format(dt, '{HH}:{mm}\n{MM}-{dd}', true)
          },
        },
      })

      const fmtShort = createShortFormatter(props.unit)
      const yAxisData = props.yAxis.map((v) => {
        return fmtShort(v[0])
      })

      conf.yAxis.push({
        type: 'category',
        data: yAxisData,
        offset: 2 * itemPadding,
        axisTick: { alignWithLabel: true },
      })

      if (props.data.length) {
        let minCount = Number.MAX_VALUE
        let maxCount = 0

        for (let point of props.data) {
          const count = point[2]
          if (count < minCount) {
            minCount = count
          }
          if (count > maxCount) {
            maxCount = count
          }
        }

        conf.visualMap = {
          show: false,
          min: minCount,
          max: maxCount,
          seriesIndex: [0],
        }
      }

      conf.series.push({
        name: 'Heatmap',
        type: 'heatmap',
        data: props.data,
        itemStyle: {
          borderColor: '#fff',
          borderWidth: itemPadding,
        },
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowColor: 'rgba(0, 0, 0, 0.5)',
          },
        },
      })
    }

    function onResize(event: any) {
      const width = event.detail.width
      itemSize.value = width > 800 ? 14 : 10
    }

    return {
      echart,
      chart,

      onResize,
    }
  },
})
</script>

<style lang="scss" scoped></style>
