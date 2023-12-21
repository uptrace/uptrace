<template>
  <div>
    <HeatmapChart
      :loading="heatmapQuery.loading"
      :resolved="heatmapQuery.status.isResolved()"
      :x-axis="heatmapQuery.xAxis"
      :y-axis="heatmapQuery.yAxis"
      :data="heatmapQuery.data"
      :unit="gridItem.params.unit"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { joinQuery, injectQueryStore } from '@/use/uql'
import { useHeatmapQuery } from '@/metrics/use-query'

// Components
import HeatmapChart from '@/components/HeatmapChart.vue'

// Misc
import { Dashboard, HeatmapGridItem } from '@/metrics/types'

export default defineComponent({
  name: 'GridItemHeatmap',
  components: { HeatmapChart },

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
      type: Object as PropType<HeatmapGridItem>,
      required: true,
    },
  },

  setup(props, ctx) {
    const { where } = injectQueryStore()
    const heatmapQuery = useHeatmapQuery(() => {
      if (!props.gridItem.params.metric) {
        return undefined
      }

      return {
        ...props.dateRange.axiosParams(),
        time_offset: props.dashboard.timeOffset,
        metric: props.gridItem.params.metric,
        alias: props.gridItem.params.metric,
        query: joinQuery([props.gridItem.params.query, where.value]),
      }
    })

    watch(
      () => heatmapQuery.status,
      (status) => {
        if (status.isResolved()) {
          ctx.emit('ready')
        }
      },
    )
    watch(
      () => heatmapQuery.error,
      (error) => {
        ctx.emit('error', error)
      },
    )

    return {
      heatmapQuery,
    }
  },
})
</script>

<style lang="scss" scoped></style>
