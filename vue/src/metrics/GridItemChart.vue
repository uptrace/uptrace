<template>
  <LegendaryChart
    :loading="timeseries.loading"
    :resolved="timeseries.status.isResolved()"
    :timeseries="styledTimeseries"
    :time="timeseries.time"
    :chart-kind="gridItem.params.chartKind"
    :legend="legend"
    :height="height"
  />
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { joinQuery, injectQueryStore } from '@/use/uql'
import { useTimeseries, useStyledTimeseries } from '@/metrics/use-query'

// Components
import LegendaryChart from '@/metrics/LegendaryChart.vue'

// Misc
import {
  defaultChartLegend,
  Dashboard,
  ChartGridItem,
  ChartLegend,
  LegendType,
  LegendPlacement,
  LegendValue,
} from '@/metrics/types'

export default defineComponent({
  name: 'GridItemChart',
  components: {
    LegendaryChart,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
    gridItem: {
      type: Object as PropType<ChartGridItem>,
      required: true,
    },
    height: {
      type: Number,
      required: true,
    },
    wide: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const legend = computed((): ChartLegend => {
      if (props.wide) {
        return {
          type: LegendType.Table,
          placement: LegendPlacement.Bottom,
          values: [LegendValue.Avg, LegendValue.Last, LegendValue.Min, LegendValue.Max],
          maxLength: 150,
        }
      }
      return props.gridItem.params.legend ?? defaultChartLegend()
    })

    const { where } = injectQueryStore()
    const timeseries = useTimeseries(() => {
      if (!props.gridItem.params.metrics.length || !props.gridItem.params.query) {
        return undefined
      }

      return {
        ...props.dateRange.axiosParams(),
        time_offset: props.dashboard.timeOffset,
        metric: props.gridItem.params.metrics.map((m) => m.name),
        alias: props.gridItem.params.metrics.map((m) => m.alias),
        query: joinQuery([props.gridItem.params.query, where.value]),
        min_interval: props.dashboard.minInterval,
      }
    })

    watch(
      () => timeseries.error,
      (error) => {
        ctx.emit('error', error)
      },
      { immediate: true },
    )

    const styledTimeseries = useStyledTimeseries(
      computed(() => timeseries.items),
      computed(() => props.gridItem.params.columnMap),
      computed(() => props.gridItem.params.timeseriesMap),
    )

    return {
      legend,
      timeseries,
      styledTimeseries,
    }
  },
})
</script>

<style lang="scss" scoped></style>
