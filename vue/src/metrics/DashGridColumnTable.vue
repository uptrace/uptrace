<template>
  <div class="px-2">
    <TimeseriesTable
      :loading="tableQuery.loading"
      :items="tableQuery.items"
      :items-per-page="5"
      :columns="tableQuery.columns"
      :order="tableQuery.order"
      :axios-params="tableQuery.axiosParams"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useTableQuery } from '@/metrics/use-query'

// Components
import TimeseriesTable from '@/metrics/TimeseriesTable.vue'

// Utilities
import { TableGridColumn } from '@/metrics/types'

export default defineComponent({
  name: 'DashGridColumnChart',
  components: {
    TimeseriesTable,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    gridColumn: {
      type: Object as PropType<TableGridColumn>,
      required: true,
    },
  },

  setup(props, ctx) {
    const tableQuery = useTableQuery(
      () => {
        if (!props.gridColumn.params.metrics.length || !props.gridColumn.params.query) {
          return undefined
        }

        return {
          ...props.dateRange.axiosParams(),
          metric: props.gridColumn.params.metrics.map((m) => m.name),
          alias: props.gridColumn.params.metrics.map((m) => m.alias),
          query: props.gridColumn.params.query,
        }
      },
      computed(() => props.gridColumn.params.columnMap),
    )

    watch(
      () => tableQuery.error,
      (error) => {
        ctx.emit('error', error)
      },
    )

    return {
      tableQuery,
    }
  },
})
</script>

<style lang="scss" scoped></style>
