<template>
  <div>
    <DashGridColumnChart
      v-if="gridColumn.type === GridColumnType.Chart"
      :date-range="dateRange"
      :dashboard="dashboard"
      :grid-column="gridColumn"
      :height="height"
      :verbose="verbose"
      @error="$emit('error', $event)"
    />
    <DashGridColumnTable
      v-if="gridColumn.type === GridColumnType.Table"
      :date-range="dateRange"
      :grid-column="gridColumn"
      :height="height"
      :verbose="verbose"
      @error="$emit('error', $event)"
    />
    <DashGridColumnHeatmap
      v-if="gridColumn.type === GridColumnType.Heatmap"
      :date-range="dateRange"
      :grid-column="gridColumn"
      :height="height"
      :verbose="verbose"
      @error="$emit('error', $event)"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { Dashboard } from '@/metrics/types'

// Components
import DashGridColumnChart from '@/metrics/DashGridColumnChart.vue'
import DashGridColumnTable from '@/metrics/DashGridColumnTable.vue'
import DashGridColumnHeatmap from '@/metrics/DashGridColumnHeatmap.vue'

// Utilities
import { GridColumn, GridColumnType } from '@/metrics/types'

export default defineComponent({
  name: 'DashGridColumnItem',
  components: {
    DashGridColumnChart,
    DashGridColumnTable,
    DashGridColumnHeatmap,
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
    gridColumn: {
      type: Object as PropType<GridColumn>,
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

  setup() {
    return { GridColumnType }
  },
})
</script>

<style lang="scss" scoped></style>
