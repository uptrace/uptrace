<template>
  <DashGrid
    :date-range="dateRange"
    :dashboard="dashboard"
    :grid-rows="gridRows"
    :grid-metrics="gridMetrics"
    :grid-query="gridQuery"
    @change="$emit('change', $event)"
  />
</template>

<script lang="ts">
import { defineComponent, onBeforeUnmount, inject, PropType, Ref } from 'vue'

// Composables
import { useSyncQueryParams } from '@/use/router'
import { UseDateRange } from '@/use/date-range'

// Components
import DashGrid from '@/metrics/DashGrid.vue'

// Misc
import { Dashboard, GridRow } from '@/metrics/types'

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
    gridRows: {
      type: Array as PropType<GridRow[]>,
      required: true,
    },
    gridMetrics: {
      type: Array as PropType<string[]>,
      required: true,
    },
    gridQuery: {
      type: String,
      default: '',
    },
  },

  setup(props, ctx) {
    const footer = inject<Ref<boolean>>('footer')!
    footer.value = false
    onBeforeUnmount(() => {
      footer.value = true
    })

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
