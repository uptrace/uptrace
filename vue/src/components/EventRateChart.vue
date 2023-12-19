<template>
  <EChart
    :loading="loading"
    :height="chart.height"
    :option="chart.option"
    :annotations="annotations"
  />
</template>

<script lang="ts">
import colors from 'vuetify/lib/util/colors'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import EChart, { EChartProps } from '@/components/EChart.vue'

// Misc
import { Unit } from '@/util/fmt'
import {
  baseChartConfig,
  addChartTooltip,
  axisLabelFormatter,
  axisPointerFormatter,
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
    countPerMin: {
      type: Array as PropType<number[]>,
      default: () => [],
    },
    group: {
      type: [String, Symbol],
      default: '',
    },
    annotations: {
      type: Array as PropType<Annotation[]>,
      default: () => [],
    },
  },

  setup(props) {
    const computedGroup = computed(() => {
      if (props.group) {
        return props.group
      }
      return Symbol()
    })

    const chart = computed((): EChartProps => {
      const conf = baseChartConfig()
      addChartTooltip(conf)

      const chart = {
        height: 160,
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
      })

      conf.dataset.push({
        source: {
          time: props.time,
          countPerMin: props.countPerMin,
        },
      })

      conf.series.push({
        name: 'events per min',
        type: 'bar',
        itemStyle: { color: colors.blue.darken1 },
        encode: { x: 'time', y: 'countPerMin' },
      })

      conf.grid.push({
        top: 15,
        left: 45,
        right: 30,
        height: 120,
      })

      return chart
    })

    return {
      computedGroup,
      chart,
    }
  },
})
</script>

<style lang="scss" scoped></style>
