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
            <TimeseriesTable
              :loading="tableQuery.loading"
              :items="tableQuery.items"
              :agg-columns="tableQuery.aggColumns"
              :grouping-columns="tableQuery.groupingColumns"
              :order="tableQuery.order"
              :axios-params="tableQuery.axiosParams"
            />
          </v-col>
        </v-row>
      </template>
    </template>
    <template #options>
      <v-container fluid>
        <v-row>
          <v-col>
            <SinglePanel title="Chart options" expanded>
              <v-text-field
                v-model="dashboard.name"
                label="Dashboard name"
                filled
                dense
                :rules="rules.name"
              />
            </SinglePanel>
          </v-col>
        </v-row>

        <v-row v-for="col in tableQuery.aggColumns" :key="col.name">
          <v-col>
            <SinglePanel :title="col.name" expanded>
              <TableColumnOptionsForm :column="dashboard.tableColumnMap[col.name]" />
            </SinglePanel>
          </v-col>
        </v-row>
      </v-container>
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
import SinglePanel from '@/components/SinglePanel.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import TimeseriesTable from '@/metrics/TimeseriesTable.vue'
import TableColumnOptionsForm from '@/metrics/TableColumnOptionsForm.vue'

// Misc
import { updateColumnMap, assignColors, emptyTableColumn, Dashboard } from '@/metrics/types'
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'DashTableForm',
  components: {
    DashTableFormPanes,
    SinglePanel,
    MetricsPicker,
    MetricsQueryBuilder,
    TimeseriesTable,
    TableColumnOptionsForm,
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
    const rules = { name: [requiredRule] }

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
      () => tableQuery.aggColumns,
      (aggColumns) => {
        updateColumnMap(props.dashboard.tableColumnMap, aggColumns, emptyTableColumn)
        assignColors(props.dashboard.tableColumnMap, aggColumns)
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
      rules,

      uql,
      activeMetrics,
      tableQuery,
    }
  },
})
</script>

<style lang="scss" scoped></style>
