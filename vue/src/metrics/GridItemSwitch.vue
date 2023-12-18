<template>
  <div v-frag>
    <GridItemGauge
      v-if="gridItem.type === GridItemType.Gauge"
      :date-range="dateRange"
      :grid-item="gridItem"
      :height="height"
      v-on="$listeners"
    />
    <GridItemCard
      v-else-if="gridItem.type === GridItemType.Chart"
      :grid-item="gridItem"
      :height="height"
      expandable
      v-on="$listeners"
    >
      <template #default="{ height, wide, on }">
        <GridItemChart
          :date-range="dateRange"
          :dashboard="dashboard"
          :grid-item="gridItem"
          :height="height"
          :wide="wide"
          v-on="{ ...$listeners, ...on }"
        />
      </template>
    </GridItemCard>
    <GridItemCard
      v-else-if="gridItem.type === GridItemType.Table"
      :grid-item="gridItem"
      :height="height"
      expandable
      v-on="$listeners"
    >
      <template #default="{ height, wide, on }">
        <GridItemTable
          :date-range="dateRange"
          :dashboard="dashboard"
          :grid-item="gridItem"
          :height="height"
          :wide="wide"
          v-on="{ ...$listeners, ...on }"
        />
      </template>
    </GridItemCard>
    <GridItemCard
      v-else-if="gridItem.type === GridItemType.Heatmap"
      :grid-item="gridItem"
      :height="height"
      expandable
      v-on="$listeners"
    >
      <template #default="{ height, wide, on }">
        <GridItemHeatmap
          :date-range="dateRange"
          :dashboard="dashboard"
          :grid-item="gridItem"
          :height="height"
          :wide="wide"
          v-on="{ ...$listeners, ...on }"
        />
      </template>
    </GridItemCard>
    <div v-else>unsupported {{ gridItem.type }}</div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import GridItemGauge from '@/metrics/GridItemGauge.vue'
import GridItemCard from '@/metrics/GridItemCard.vue'
import GridItemChart from '@/metrics/GridItemChart.vue'
import GridItemTable from '@/metrics/GridItemTable.vue'
import GridItemHeatmap from '@/metrics/GridItemHeatmap.vue'

// Misc
import { Dashboard, GridItem, GridItemType } from '@/metrics/types'

export default defineComponent({
  name: 'GridItemSwitch',
  components: {
    GridItemGauge,
    GridItemCard,
    GridItemChart,
    GridItemTable,
    GridItemHeatmap,
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
      type: Object as PropType<GridItem>,
      required: true,
    },
    height: {
      type: Number,
      required: true,
    },
  },

  setup() {
    return { GridItemType }
  },
})
</script>

<style lang="scss" scoped></style>
