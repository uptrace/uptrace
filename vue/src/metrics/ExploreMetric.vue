<template>
  <v-card outlined>
    <v-toolbar color="bg--light-primary" flat>
      <v-toolbar-title>Metric explorer</v-toolbar-title>

      <v-spacer />

      <DateRangePicker :date-range="dateRange" :range-days="90" />

      <v-toolbar-items class="ml-4">
        <v-btn icon @click="$emit('click:close')">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-toolbar-items>
    </v-toolbar>

    <v-container fluid class="pa-4">
      <v-row>
        <v-col>
          <MetricsPicker v-model="metricAliases" :uql="uql" />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-card outlined rounded="lg" class="pa-4">
            <MetricsQueryBuilder
              :date-range="dateRange"
              :metrics="activeMetrics"
              :uql="uql"
              show-agg
              show-group-by
              show-dash-where
              :disabled="!activeMetrics.length"
            />
          </v-card>
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <v-divider />
        </v-col>
      </v-row>

      <v-row class="mb-n6">
        <v-col cols="auto">
          <v-chip v-for="(col, colName) in columnMap" :key="colName" outlined label class="ma-1">
            <span>{{ colName }}</span>
            <UnitPicker v-model="col.unit" target-class="mr-n4" />
          </v-chip>
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <LegendaryChart
            :loading="timeseries.loading"
            :resolved="timeseries.status.isResolved()"
            :timeseries="styledTimeseries"
            :time="timeseries.time"
            :legend="legend"
            :height="400"
          />
        </v-col>
      </v-row>
    </v-container>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, shallowRef, ref, computed, watch, PropType } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import {
  useActiveMetrics,
  defaultMetricAlias,
  defaultMetricQuery,
  ExploredMetric,
} from '@/metrics/use-metrics'
import { useTimeseries, useStyledTimeseries } from '@/metrics/use-query'
import { MetricAlias, LegendType, LegendPlacement, LegendValue } from '@/metrics/types'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import UnitPicker from '@/components/UnitPicker.vue'
import LegendaryChart from '@/metrics/LegendaryChart.vue'

// Misc
import { updateColumnMap, assignColors, emptyMetricColumn, MetricColumn } from '@/metrics/types'

export default defineComponent({
  name: 'ExploreMetric',
  components: { DateRangePicker, MetricsPicker, MetricsQueryBuilder, UnitPicker, LegendaryChart },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    metric: {
      type: Object as PropType<ExploredMetric>,
      required: true,
    },
  },

  setup(props) {
    useTitle('Explore Metrics')
    const uql = useUql('')

    const metricAliases = shallowRef<MetricAlias[]>([])
    watch(
      () => props.metric,
      (metric) => {
        const alias = defaultMetricAlias(metric.name)
        metricAliases.value = [
          {
            name: metric.name,
            alias,
          },
        ]
        const column = defaultMetricQuery(metric.instrument, alias)
        uql.query = `${column}`
      },
      { immediate: true },
    )

    const legend = computed(() => {
      return {
        type: LegendType.Table,
        placement: LegendPlacement.Bottom,
        values: [LegendValue.Avg, LegendValue.Last, LegendValue.Min, LegendValue.Max],
        maxLength: 150,
      }
    })

    const activeMetrics = useActiveMetrics(metricAliases)

    const timeseries = useTimeseries(() => {
      if (!metricAliases.value.length || !uql.query) {
        return undefined
      }

      return {
        ...props.dateRange.axiosParams(),
        metric: metricAliases.value.map((m) => m.name),
        alias: metricAliases.value.map((m) => m.alias),
        query: uql.query,
      }
    })

    const columnMap = ref<Record<string, MetricColumn>>({})
    const styledTimeseries = useStyledTimeseries(
      computed(() => timeseries.items),
      columnMap,
      computed(() => ({})),
    )

    watch(
      () => timeseries.query,
      (query) => {
        if (query) {
          uql.setQueryInfo(query)
        }
      },
      { immediate: true },
    )

    watch(
      () => timeseries.columns,
      (columns) => {
        updateColumnMap(columnMap.value, columns, emptyMetricColumn)
        assignColors(columnMap.value, columns)
      },
    )

    return { uql, legend, metricAliases, activeMetrics, columnMap, timeseries, styledTimeseries }
  },
})
</script>

<style lang="scss" scoped></style>
