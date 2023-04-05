<template>
  <GridColumnChart
    :loading="timeseries.loading"
    :resolved="timeseries.status.isResolved()"
    :timeseries="styledTimeseries"
    :time="timeseries.time"
    :chart-kind="gridColumn.params.chartKind"
    :legend="legend"
    :height="height"
  />
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useTimeseries, useStyledTimeseries } from '@/metrics/use-query'

// Components
import GridColumnChart from '@/metrics/GridColumnChart.vue'

// Utilities
import {
  ChartGridColumn,
  ChartLegend,
  LegendType,
  LegendPlacement,
  LegendValue,
} from '@/metrics/types'

export default defineComponent({
  name: 'DashGridColumnChart',
  components: {
    GridColumnChart,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    gridColumn: {
      type: Object as PropType<ChartGridColumn>,
      required: true,
    },
    height: {
      type: Number,
      required: true,
    },
    verbose: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const legend = computed((): ChartLegend => {
      if (props.verbose) {
        return {
          type: LegendType.Table,
          placement: LegendPlacement.Bottom,
          values: [LegendValue.Avg, LegendValue.Last, LegendValue.Min, LegendValue.Max],
          maxLength: 150,
        }
      }
      return props.gridColumn.params.legend
    })

    const timeseries = useTimeseries(() => {
      if (!props.gridColumn.params.metrics.length || !props.gridColumn.params.query) {
        return undefined
      }

      return {
        ...props.dateRange.axiosParams(),
        metric: props.gridColumn.params.metrics.map((m) => m.name),
        alias: props.gridColumn.params.metrics.map((m) => m.alias),
        query: props.gridColumn.params.query,
      }
    })

    watch(
      () => timeseries.error,
      (error) => {
        if (error) {
          ctx.emit('error', error)
        }
      },
    )

    const styledTimeseries = useStyledTimeseries(
      computed(() => timeseries.items),
      computed(() => props.gridColumn.params.columnMap),
      computed(() => props.gridColumn.params.timeseriesMap),
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
