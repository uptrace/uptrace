<template>
  <div>
    <GridItemGaugeForm
      v-if="gridItem.type === GridItemType.Gauge"
      v-bind="attrs"
      v-on="$listeners"
    />
    <GridItemChartForm
      v-else-if="gridItem.type === GridItemType.Chart"
      v-bind="attrs"
      v-on="$listeners"
    />
    <GridItemTableForm
      v-else-if="gridItem.type === GridItemType.Table"
      v-bind="attrs"
      v-on="$listeners"
    />
    <GridItemHeatmapForm
      v-else-if="gridItem.type === GridItemType.Heatmap"
      v-bind="attrs"
      v-on="$listeners"
    />
    <div v-else>unsupported {{ gridItem.type }}</div>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { provideForceReload } from '@/use/force-reload'
import { useDateRangeFrom, UseDateRange } from '@/use/date-range'

// Components
import GridItemGaugeForm from '@/metrics/GridItemGaugeForm.vue'
import GridItemChartForm from '@/metrics/GridItemChartForm.vue'
import GridItemTableForm from '@/metrics/GridItemTableForm.vue'
import GridItemHeatmapForm from '@/metrics/GridItemHeatmapForm.vue'

// Misc
import { Dashboard, GridItem, GridItemType } from '@/metrics/types'

export default defineComponent({
  name: 'GridItemFormSwitch',
  components: {
    GridItemGaugeForm,
    GridItemChartForm,
    GridItemTableForm,
    GridItemHeatmapForm,
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
    tableGrouping: {
      type: Array as PropType<string[]>,
      required: true,
    },
    gridItem: {
      type: Object as PropType<GridItem>,
      required: true,
    },
  },

  setup(props) {
    provideForceReload()
    const internalDateRange = useDateRangeFrom(props.dateRange)
    const attrs = computed(() => {
      return {
        dateRange: internalDateRange,
        dashboard: props.dashboard,
        tableGrouping: props.tableGrouping,
        gridItem: props.gridItem,
      }
    })
    return { GridItemType, attrs }
  },
})
</script>

<style lang="scss" scoped></style>
