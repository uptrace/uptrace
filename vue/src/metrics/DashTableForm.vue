<template>
  <DashTableFormPanes :dashboard="dashboard" v-on="$listeners">
    <template #picker>
      <MetricsPicker v-model="dashboard.tableMetrics" :uql="uql" auto-grouping />
    </template>
    <template #preview>
      <v-row v-if="!activeMetrics.length">
        <v-col>
          <v-skeleton-loader type="table" :boilerplate="!tableQuery.loading"></v-skeleton-loader>
        </v-col>
      </v-row>

      <template v-else>
        <v-row>
          <v-col>
            <MetricsQueryBuilder
              :date-range="dateRange"
              :metrics="activeMetrics"
              :uql="uql"
              show-agg
              show-group-by
              show-dash-where
              :disabled="!activeMetrics.length"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-row dense>
              <v-col v-for="(col, colName) in dashboard.tableColumnMap" :key="colName" cols="auto">
                <MetricColumnChip :name="colName" :column="col" />
              </v-col>
            </v-row>

            <v-row dense>
              <v-col>
                <TimeseriesTable
                  :loading="tableQuery.loading"
                  :items="tableQuery.items"
                  :columns="tableQuery.columns"
                  :order="tableQuery.order"
                  :axios-params="tableQuery.axiosParams"
                />
              </v-col>
            </v-row>
          </v-col>
        </v-row>
      </template>
    </template>
  </DashTableFormPanes>
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import { useActiveMetrics } from '@/metrics/use-metrics'
import { useTableQuery } from '@/metrics/use-query'

// Components
import DashTableFormPanes from '@/metrics/DashTableFormPanes.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import TimeseriesTable from '@/metrics/TimeseriesTable.vue'
import MetricColumnChip from '@/metrics/MetricColumnChip.vue'

// Misc
import { updateColumnMap, Dashboard, MetricColumn } from '@/metrics/types'
import { eChart as colorScheme } from '@/util/colorscheme'

export default defineComponent({
  name: 'DashTableForm',
  components: {
    DashTableFormPanes,
    MetricsPicker,
    MetricsQueryBuilder,
    TimeseriesTable,
    MetricColumnChip,
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
  },

  setup(props, ctx) {
    const uql = useUql()
    const activeMetrics = useActiveMetrics(computed(() => props.dashboard.tableMetrics))

    const tableQuery = useTableQuery(
      () => {
        if (!props.dashboard.tableQuery || !props.dashboard.tableMetrics.length) {
          return undefined
        }

        return {
          ...props.dateRange.axiosParams(),
          metric: props.dashboard.tableMetrics.map((m) => m.name),
          alias: props.dashboard.tableMetrics.map((m) => m.alias),
          query: props.dashboard.tableQuery,
        }
      },
      computed(() => props.dashboard.tableColumnMap),
    )

    watch(
      () => tableQuery.columns,
      (columns) => {
        updateColumnMap(props.dashboard.tableColumnMap, columns)
        assignColors(props.dashboard.tableColumnMap)
      },
    )

    watch(
      () => props.dashboard.tableQuery,
      (query) => {
        uql.query = query
      },
      { immediate: true },
    )

    watch(
      () => uql.query,
      (query) => {
        props.dashboard.tableQuery = query
      },
    )
    watch(
      () => tableQuery.query,
      (query) => {
        if (query) {
          uql.setQueryInfo(query)
        }
      },
      { immediate: true },
    )

    return {
      uql,

      activeMetrics,
      tableQuery,
    }
  },
})

function assignColors(colMap: Record<string, MetricColumn>) {
  const colors = new Set(colorScheme)

  for (let colName in colMap) {
    const col = colMap[colName]
    if (!col.color) {
      continue
    }
    colors.delete(col.color)
  }

  const values = colors.values()
  for (let colName in colMap) {
    const col = colMap[colName]
    if (col.color) {
      continue
    }

    const first = values.next()
    col.color = first.value
  }
}
</script>

<style lang="scss" scoped></style>
