<template>
  <div class="px-2">
    <TimeseriesTable
      :loading="tableQuery.loading"
      :items="tableQuery.items"
      :items-per-page="gridItem.params.itemsPerPage"
      :agg-columns="tableQuery.aggColumns"
      :grouping-columns="tableQuery.groupingColumns"
      :order="tableQuery.order"
      :axios-params="tableQuery.axiosParams"
      :dense="gridItem.params.denseTable"
      @current-items="$emit('ready')"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { joinQuery, injectQueryStore } from '@/use/uql'
import { useTableQuery } from '@/metrics/use-query'

// Components
import TimeseriesTable from '@/metrics/TimeseriesTable.vue'

// Misc
import { Dashboard, TableGridItem } from '@/metrics/types'

export default defineComponent({
  name: 'GridItemTable',
  components: {
    TimeseriesTable,
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
      type: Object as PropType<TableGridItem>,
      required: true,
    },
  },

  setup(props, ctx) {
    const { where } = injectQueryStore()
    const tableQuery = useTableQuery(
      () => {
        if (!props.gridItem.params.metrics.length || !props.gridItem.params.query) {
          return undefined
        }

        return {
          ...props.dateRange.axiosParams(),
          time_offset: props.dashboard.timeOffset,
          metric: props.gridItem.params.metrics.map((m) => m.name),
          alias: props.gridItem.params.metrics.map((m) => m.alias),
          query: joinQuery([props.gridItem.params.query, where.value]),
        }
      },
      computed(() => props.gridItem.params.columnMap),
    )

    watch(
      () => tableQuery.status,
      (status) => {
        if (status.isResolved()) {
          ctx.emit('ready')
        }
      },
    )
    watch(
      () => tableQuery.queryError,
      (error) => {
        ctx.emit('error', error)
      },
    )

    watch(
      () => tableQuery.query,
      (query) => {
        if (query) {
          ctx.emit('update:query', query)
        }
      },
    )
    watch(
      () => tableQuery.columns,
      (columns) => {
        ctx.emit('update:columns', columns)
      },
    )

    return {
      tableQuery,
    }
  },
})
</script>

<style lang="scss" scoped></style>
