<template>
  <div>
    <v-row class="px-2 text-subtitle-1">
      <v-col>
        <UptraceQueryChip
          v-for="(part, i) in queryParts"
          :key="i"
          :query="part.query"
          class="mr-2 mb-1"
        />
      </v-col>
    </v-row>

    <v-row align="end" class="px-2 text-subtitle-2 text-center">
      <v-col cols="auto">
        <div class="grey--text font-weight-regular">State</div>
        <div>{{ alert.state }}</div>
      </v-col>

      <v-col cols="auto">
        <div class="grey--text font-weight-regular">Trigger</div>
        <MetricMonitorTrigger :alert="alert" />
      </v-col>

      <v-col cols="auto">
        <div class="grey--text font-weight-regular">Time</div>
        <XDate :date="alert.updatedAt" />
      </v-col>

      <v-col v-if="alert.createdAt !== alert.updatedAt" cols="auto">
        <div class="grey--text font-weight-regular">First seen</div>
        <XDate :date="alert.createdAt" />
      </v-col>

      <v-col cols="auto">
        <v-btn
          depressed
          small
          :to="{ name: 'MonitorMetricShow', params: { monitorId: alert.monitorId } }"
          exact
          >View monitor</v-btn
        >
        <v-btn v-if="spansRoute" :to="spansRoute" depressed small class="ml-2"
          >Matching spans</v-btn
        >
        <slot v-if="$slots['append-action']" name="append-action" />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <v-card outlined rounded="lg" class="pa-4">
          <MetricChart
            :loading="timeseries.loading"
            :resolved="timeseries.status.isResolved()"
            :timeseries="styledTimeseries"
            :time="timeseries.time"
            :height="300"
            :min-allowed-value="alert.params.bounds.min"
            :max-allowed-value="alert.params.bounds.max"
            show-legend
          />
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import colors from 'vuetify/lib/util/colors'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { parseParts } from '@/use/uql'
import { UseDateRange } from '@/use/date-range'
import { useTimeseries, useStyledTimeseries } from '@/metrics/use-query'
import { MetricAlert, AlertState } from '@/alerting/use-alerts'

// Components
import UptraceQueryChip from '@/components/UptraceQueryChip.vue'
import MetricChart from '@/metrics/MetricChart.vue'
import MetricMonitorTrigger from '@/alerting/MetricMonitorTrigger.vue'

// Utils
import { MetricColumn } from '@/metrics/types'
import { SystemName, AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'MetricAlertCard',
  components: {
    UptraceQueryChip,
    MetricChart,
    MetricMonitorTrigger,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    alert: {
      type: Object as PropType<MetricAlert>,
      required: true,
    },
  },

  setup(props, ctx) {
    const monitor = computed(() => {
      return props.alert.params.monitor
    })

    const columnMap = computed((): Record<string, MetricColumn> => {
      return {
        [monitor.value.column]: {
          unit: monitor.value.columnUnit,
          color: colors.blue.lighten1,
        },
      }
    })

    const timeseries = useTimeseries(() => {
      return {
        ...props.dateRange.axiosParams(),
        metric: monitor.value.metrics.map((m) => m.name),
        alias: monitor.value.metrics.map((m) => m.alias),
        query: monitor.value.query,
      }
    })
    const styledTimeseries = useStyledTimeseries(
      computed(() => timeseries.items),
      columnMap,
      computed(() => ({})),
    )

    const queryParts = computed(() => {
      const tables = []
      for (let metric of monitor.value.metrics) {
        tables.push(`${metric.name} as $${metric.alias}`)
      }

      return parseParts(`${monitor.value.query} | from ${tables.join(', ')}`)
    })

    const spansRoute = computed(() => {
      const where = monitor.value.query.split(' | ').filter((part) => part.startsWith('where '))
      if (!where.length) {
        return undefined
      }

      return {
        name: 'SpanList',
        query: {
          system: SystemName.All,
          query: [
            AttrKey.spanCountPerMin,
            AttrKey.spanErrorRate,
            `{p50,p90,p99}(${AttrKey.spanDuration})`,
            ...where,
          ].join(' | '),
        },
      }
    })

    return {
      AlertState,

      queryParts,
      timeseries,
      styledTimeseries,
      spansRoute,
    }
  },
})
</script>

<style lang="scss" scoped></style>
