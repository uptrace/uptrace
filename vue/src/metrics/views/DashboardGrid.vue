<template>
  <DashGrid
    :date-range="dateRange"
    :dashboard="dashboard"
    :grid="grid"
    :grid-query="gridQuery"
    :editable="editable"
    @change="$emit('change', $event)"
  />
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useSyncQueryParams } from '@/use/router'
import { UseDateRange } from '@/use/date-range'

// Components
import DashGrid from '@/metrics/DashGrid.vue'

// Utilities
import { Dashboard, GridColumn } from '@/metrics/types'

export default defineComponent({
  name: 'DashboardGrid',
  components: { DashGrid },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
    grid: {
      type: Array as PropType<GridColumn[]>,
      required: true,
    },
    gridQuery: {
      type: String,
      default: '',
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    useSyncQueryParams({
      fromQuery(queryParams) {
        props.dateRange.parseQueryParams(queryParams)
      },
      toQuery() {
        return {
          ...props.dateRange.queryParams(),
        }
      },
    })

    return {}
  },
})
</script>

<style lang="scss" scoped></style>
