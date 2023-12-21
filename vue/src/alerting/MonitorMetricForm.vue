<template>
  <MonitorMetricFormPanes :date-range="dateRange" :monitor="monitor" v-on="$listeners">
    <template #title>
      <slot name="title"></slot>
    </template>
    <template #picker>
      <MetricsPicker v-model="monitor.params.metrics" :metrics="metrics" :uql="uql" editable />
    </template>
    <template #preview>
      <v-container fluild>
        <v-row v-if="!activeMetrics.length">
          <v-col>
            <v-skeleton-loader type="image" boilerplate></v-skeleton-loader>
          </v-col>
        </v-row>

        <template v-else>
          <v-row>
            <v-col>
              <MetricsQueryBuilder
                :date-range="dateRange"
                :metrics="activeMetrics"
                :uql="uql"
                :disabled="!activeMetrics.length"
                show-agg
                show-group-by
                show-dash-where
              />

              <div
                v-if="Object.keys(internalColumnMap).length > 1"
                class="mt-1 d-flex align-center"
              >
                <div>
                  <v-icon size="30" color="red darken-1" class="mr-3">mdi-alert-circle</v-icon>
                </div>
                <div class="text-body-2">
                  The query returns {{ Object.keys(internalColumnMap).length }} columns, but only a
                  single column is allowed.<br />
                  To keep the column but hide the result, underscore the alias, for example,
                  <code>count($metric) as _tmp_count</code>.
                </div>
              </div>
              <div
                v-else-if="
                  timeseries.status.hasData() && Object.keys(internalColumnMap).length === 0
                "
                class="mt-1 d-flex align-center"
              >
                <div>
                  <v-icon size="30" color="red darken-1" class="mr-3">mdi-alert-circle</v-icon>
                </div>
                <div class="text-body-2">The query must return at least one column to monitor.</div>
              </div>
            </v-col>
          </v-row>

          <v-row class="mb-n6">
            <v-col>
              <v-chip
                v-for="(col, colName) in internalColumnMap"
                :key="colName"
                outlined
                label
                class="ma-1"
              >
                <span>{{ colName }}</span>
                <UnitPicker v-model="col.unit" target-class="mr-n4" />
              </v-chip>
            </v-col>
          </v-row>

          <v-row>
            <v-col>
              <MetricChart
                :loading="timeseries.loading"
                :resolved="timeseries.status.isResolved()"
                :timeseries="styledTimeseries"
                :time="timeseries.time"
                :min-allowed-value="monitor.params.minValue"
                :max-allowed-value="monitor.params.maxValue"
                :event-bus="eventBus"
              />
            </v-col>
          </v-row>

          <v-row v-if="timeseries.items.length" no-gutters justify="center">
            <v-col cols="auto">
              <ChartLegendTable
                :timeseries="styledTimeseries"
                @hover:item="eventBus.emit('hover', $event)"
              />
            </v-col>
          </v-row>
        </template>
      </v-container>
    </template>
    <template #options="{ form }">
      <MonitorMetricFormOptions
        :monitor="monitor"
        :column-map="internalColumnMap"
        :timeseries="timeseries.items"
        :form="form"
      />
    </template>
  </MonitorMetricFormPanes>
</template>

<script lang="ts">
import { defineComponent, ref, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import { useActiveMetrics } from '@/metrics/use-metrics'
import { useTimeseries, useStyledTimeseries } from '@/metrics/use-query'

// Components
import UnitPicker from '@/components/UnitPicker.vue'
import MonitorMetricFormPanes from '@/alerting/MonitorMetricFormPanes.vue'
import MonitorMetricFormOptions from '@/alerting/MonitorMetricFormOptions.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import MetricChart from '@/metrics/MetricChart.vue'
import ChartLegendTable from '@/metrics/ChartLegendTable.vue'

// Misc
import { EventBus } from '@/models/eventbus'
import { updateColumnMap, emptyMetricColumn, Metric, MetricColumn } from '@/metrics/types'
import { MetricMonitor } from '@/alerting/types'

export default defineComponent({
  name: 'MonitorMetricForm',
  components: {
    MonitorMetricFormPanes,
    MonitorMetricFormOptions,
    UnitPicker,
    MetricsPicker,
    MetricsQueryBuilder,
    MetricChart,
    ChartLegendTable,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    metrics: {
      type: Array as PropType<Metric[]>,
      required: true,
    },
    monitor: {
      type: Object as PropType<MetricMonitor>,
      required: true,
    },
    columnMap: {
      type: Object as PropType<Record<string, MetricColumn>>,
      default: undefined,
    },
  },

  setup(props, ctx) {
    const eventBus = new EventBus()

    const internalColumnMap = ref<Record<string, MetricColumn>>({})
    watch(
      () => props.columnMap,
      (columnMap) => {
        if (!columnMap) {
          return
        }
        for (let key in columnMap) {
          internalColumnMap.value[key] = columnMap[key]
        }
      },
      { immediate: true },
    )

    const uql = useUql()
    const activeMetrics = useActiveMetrics(computed(() => props.monitor.params.metrics))
    const axiosParams = computed(() => {
      if (!props.monitor.params.query) {
        return undefined
      }

      const metrics = props.monitor.params.metrics.filter((m) => m.name && m.alias)
      if (!metrics.length) {
        return undefined
      }

      return {
        ...props.dateRange.axiosParams(),
        time_offset: props.monitor.params.timeOffset,

        metric: metrics.map((m) => m.name),
        alias: metrics.map((m) => m.alias),
        query: props.monitor.params.query,
      }
    })

    const timeseries = useTimeseries(() => {
      return axiosParams.value
    })

    const styledTimeseries = useStyledTimeseries(
      computed(() => timeseries.items),
      computed(() => internalColumnMap.value),
      computed(() => ({})),
    )

    watch(
      () => props.monitor.params.query,
      (query) => {
        uql.query = query
      },
      { immediate: true },
    )
    watch(
      () => uql.query,
      (query) => {
        props.monitor.params.query = query
      },
    )

    watch(
      () => timeseries.query,
      (queryInfo) => {
        if (queryInfo) {
          uql.setQueryInfo(queryInfo)
        }
      },
      { immediate: true },
    )
    watch(
      () => timeseries.columns,
      (columns) => {
        updateColumnMap(internalColumnMap.value, columns, emptyMetricColumn)

        const params = props.monitor.params
        if (params.column in internalColumnMap.value) {
          internalColumnMap.value[params.column].unit = params.columnUnit
        }
      },
    )

    return {
      internalColumnMap,
      eventBus,

      uql,
      activeMetrics,
      axiosParams,
      timeseries,
      styledTimeseries,
    }
  },
})
</script>

<style lang="scss" scoped></style>
