<template>
  <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="py-4">
    <v-card outlined>
      <v-toolbar color="light-blue lighten-5" flat>
        <v-toolbar-title>Quickly Test Metrics</v-toolbar-title>

        <v-spacer />

        <DateRangePicker :date-range="dateRange" :range-days="90" />
      </v-toolbar>

      <v-container class="py-6 indent-rows">
        <v-row align="center" dense>
          <v-col cols="auto" class="pr-4">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">1</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="text-subtitle-1 text--primary">
              Select up to 5 metrics you want to display for each row in the table. The selected
              metrics should have some common attributes that will be used to join metrics together.
            </v-sheet>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricsPicker v-model="metricAliases" :metrics="metrics.items" :uql="uql" />
          </v-col>
        </v-row>

        <v-divider class="my-8" />

        <v-row align="center" dense>
          <v-col cols="auto" class="pr-4">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">2</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="text-subtitle-1 text--primary">
              Select some aggregations and group-by attributes to display as columns in the table.
            </v-sheet>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricQueryBuilder
              :date-range="dateRange"
              :metrics="activeMetrics"
              :uql="uql"
              show-dash-group-by
              :disabled="!activeMetrics.length"
            />
          </v-col>
        </v-row>

        <v-divider class="my-8" />

        <v-row dense>
          <v-col cols="auto">
            <v-chip v-for="(col, colName) in columnMap" :key="colName" outlined label class="ma-1">
              <span>{{ colName }}</span>
              <UnitPicker v-model="col.unit" target-class="mr-n4" />
            </v-chip>
          </v-col>
        </v-row>
        <v-row dense>
          <v-col>
            <MetricItemsTable
              :loading="tableQuery.loading"
              :items="tableQuery.items"
              :columns="tableQuery.columns"
              :order="tableQuery.order"
              :axios-params="axiosParams"
              :column-map="columnMap"
            />
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, shallowRef, reactive, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import { useActiveMetrics, UseMetrics, MetricAlias } from '@/metrics/use-metrics'
import { useTableQuery } from '@/metrics/use-query'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricQueryBuilder from '@/metrics/query/MetricQueryBuilder.vue'
import MetricItemsTable from '@/metrics/MetricItemsTable.vue'
import UnitPicker from '@/components/UnitPicker.vue'

// Types
import { MetricColumn } from '@/metrics/types'

export default defineComponent({
  name: 'MetricsExplore',
  components: { DateRangePicker, MetricsPicker, MetricQueryBuilder, MetricItemsTable, UnitPicker },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    metrics: {
      type: Object as PropType<UseMetrics>,
      required: true,
    },
  },

  setup(props) {
    const uql = useUql({ query: 'group by all' })
    const metricAliases = shallowRef<MetricAlias[]>([])

    const activeMetrics = useActiveMetrics(
      computed(() => props.metrics.items),
      metricAliases,
    )

    const axiosParams = computed(() => {
      if (!metricAliases.value.length || !uql.query) {
        return { _: undefined }
      }

      return {
        ...props.dateRange.axiosParams(),
        metrics: metricAliases.value.map((m) => m.name),
        aliases: metricAliases.value.map((m) => m.alias),
        query: uql.query,
      }
    })

    const tableQuery = useTableQuery(axiosParams)

    const columnMap = computed((): Record<string, MetricColumn> => {
      const columnMap = {}

      for (let col of tableQuery.columns) {
        if (!col.isGroup) {
          columnMap[col.name] = {
            unit: col.unit,
          }
        }
      }

      return reactive(columnMap)
    })

    watch(
      () => tableQuery.queryParts,
      (queryParts) => {
        if (queryParts) {
          uql.syncParts(queryParts)
        }
      },
      { immediate: true },
    )

    return { uql, metricAliases, activeMetrics, axiosParams, tableQuery, columnMap }
  },
})
</script>

<style lang="scss" scoped>
.indent-rows ::v-deep .row {
  padding-left: 12px !important;
  padding-right: 12px !important;
}
</style>
