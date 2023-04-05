<template>
  <HeatmapChart
    :loading="heatmapQuery.loading"
    :resolved="heatmapQuery.status.isResolved()"
    :unit="gridColumn.params.unit"
    :x-axis="heatmapQuery.xAxis"
    :y-axis="heatmapQuery.yAxis"
    :data="heatmapQuery.data"
  />
</template>

<script lang="ts">
import { defineComponent, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useHeatmapQuery } from '@/metrics/use-query'

// Components
import HeatmapChart from '@/components/HeatmapChart.vue'

// Utilities
import { HeatmapGridColumn } from '@/metrics/types'

export default defineComponent({
  name: 'DashGridColumnHeatmap',
  components: { HeatmapChart },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    gridColumn: {
      type: Object as PropType<HeatmapGridColumn>,
      required: true,
    },
  },

  setup(props, ctx) {
    const heatmapQuery = useHeatmapQuery(() => {
      if (!props.gridColumn.params.metric) {
        return undefined
      }

      return {
        ...props.dateRange.axiosParams(),
        metric: props.gridColumn.params.metric,
        alias: props.gridColumn.params.metric,
        query: props.gridColumn.params.query,
      }
    })

    watch(
      () => heatmapQuery.error,
      (error) => {
        if (error) {
          ctx.emit('error', error)
        }
      },
    )

    return {
      heatmapQuery,
    }
  },
})
</script>

<style lang="scss" scoped></style>
