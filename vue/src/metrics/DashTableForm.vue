<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-card outlined>
      <v-toolbar color="light-blue lighten-5" flat>
        <v-toolbar-title>Dashboard table</v-toolbar-title>
        <v-btn icon href="https://uptrace.dev/get/querying-metrics.html" target="_blank"
          ><v-icon>mdi-help-circle-outline</v-icon></v-btn
        >

        <v-spacer />

        <v-btn small outlined :loading="tableQuery.loading" @click="tableQuery.reload()">
          <v-icon small left>mdi-refresh</v-icon>
          <span>Reload</span>
        </v-btn>

        <v-toolbar-items class="ml-4">
          <v-btn icon @click="$emit('click:cancel')">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>
      </v-toolbar>

      <v-container fluid class="pa-6">
        <v-row align="center" dense>
          <v-col cols="auto">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">1</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="px-4 text-subtitle-1 text--primary">
              Select metrics you want to display for each row in the table. The selected metrics
              should have some common attributes that will be used to join timeseries together.
            </v-sheet>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricsPicker v-model="dashboard.tableMetrics" :uql="uql" :editable="editable" />
          </v-col>
        </v-row>

        <v-divider class="my-8" />

        <v-row align="center" dense>
          <v-col cols="auto">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">2</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="px-4 text-subtitle-1 text--primary">
              Add some aggregations and group-by attributes to display as columns in the table. Each
              row in the table will lead to a separate grid-based dashboard.
            </v-sheet>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricsQueryBuilder
              :date-range="dateRange"
              :metrics="activeMetrics"
              :uql="uql"
              show-agg
              show-group-by
              show-metrics-where
              :disabled="!activeMetrics.length"
            />
          </v-col>
        </v-row>

        <v-divider class="my-8" />

        <v-row>
          <v-col>
            <v-row dense>
              <v-col v-for="(col, colName) in dashboard.tableColumnMap" :key="colName" cols="auto">
                <MetricColumnChip :name="colName" :column="col" />
              </v-col>
            </v-row>

            <v-row>
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

        <v-row v-if="editable" class="mt-8">
          <v-col>
            <v-divider />
          </v-col>
        </v-row>

        <v-row v-if="editable">
          <v-spacer />
          <v-col cols="auto">
            <v-btn text class="mr-2" @click="$emit('click:cancel')">Cancel</v-btn>
            <v-btn type="submit" color="primary" :disabled="!isValid" :loading="dashMan.pending"
              >Save</v-btn
            >
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import { useActiveMetrics } from '@/metrics/use-metrics'
import { useDashManager } from '@/metrics/use-dashboards'
import { useTableQuery } from '@/metrics/use-query'

// Components
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import TimeseriesTable from '@/metrics/TimeseriesTable.vue'
import MetricColumnChip from '@/metrics/MetricColumnChip.vue'

// Utilities
import { updateColumnMap, Dashboard, MetricColumn } from '@/metrics/types'
import { eChart as colorScheme } from '@/util/colorscheme'

export default defineComponent({
  name: 'DashTableForm',
  components: {
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
    editable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const uql = useUql()

    const isValid = shallowRef(false)
    const dashMan = useDashManager()

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

    function submit() {
      dashMan
        .updateTable({
          tableMetrics: props.dashboard.tableMetrics,
          tableQuery: props.dashboard.tableQuery,
          tableColumnMap: props.dashboard.tableColumnMap,
        })
        .then((dash) => {
          ctx.emit('click:save', dash)
        })
    }

    return {
      uql,

      isValid,
      dashMan,
      submit,

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
