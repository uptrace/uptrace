<template>
  <div>
    <v-row class="px-2 text-subtitle-1">
      <v-col>
        <UqlCardReadonly :query="query" />
      </v-col>
    </v-row>

    <v-row align="end" class="px-2 text-subtitle-2 text-center">
      <v-col cols="auto">
        <div class="grey--text font-weight-regular">Status</div>
        <div>{{ alert.status }}</div>
      </v-col>

      <v-col cols="auto">
        <div class="grey--text font-weight-regular">Trigger</div>
        <MetricMonitorTrigger :alert="alert" />
      </v-col>

      <v-col cols="auto">
        <div class="grey--text font-weight-regular">Time</div>
        <DateValue :value="alert.time" />
      </v-col>

      <v-col cols="auto">
        <v-btn
          depressed
          small
          :to="{ name: 'MonitorMetricShow', params: { monitorId: alert.monitorId } }"
          exact
          >View monitor</v-btn
        >
        <v-btn v-if="routeForSpans" :to="routeForSpans" depressed small class="ml-2"
          >Matching spans</v-btn
        >
        <slot v-if="$slots['append-action']" name="append-action" />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        You specified that value should be between <strong>{{ minValue }}</strong> and
        <strong>{{ maxValue }}</strong
        >. The actual value of <strong>{{ currentValue }}</strong> ({{ currentValueVerbose }}) has
        been {{ params.firing === -1 ? 'smaller' : 'greater' }} than this range for at least
        <strong>{{ duration }}</strong
        >.
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
            :mark-point="markPoint"
            show-legend
          />
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import colors from 'vuetify/lib/util/colors'
import { formatDuration } from 'date-fns'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { joinQuery } from '@/use/uql'
import { UseDateRange } from '@/use/date-range'
import { useTimeseries, useStyledTimeseries } from '@/metrics/use-query'
import { MetricAlert, AlertStatus } from '@/alerting/use-alerts'

// Components
import UqlCardReadonly from '@/components/UqlCardReadonly.vue'
import MetricChart from '@/metrics/MetricChart.vue'
import MetricMonitorTrigger from '@/alerting/MetricMonitorTrigger.vue'

// Utils
import { MetricColumn } from '@/metrics/types'
import { SystemName, AttrKey } from '@/models/otel'
import { fmt, numVerbose } from '@/util/fmt'
import { MINUTE, HOUR } from '@/util/fmt/date'

export default defineComponent({
  name: 'MetricAlertCard',
  components: {
    UqlCardReadonly,
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
    const params = computed(() => {
      return props.alert.params
    })
    const monitor = computed(() => {
      return props.alert.params.monitor
    })
    const minValue = computed(() => {
      if (params.value.bounds.min === null) {
        return '-Inf'
      }
      return fmt(params.value.bounds.min, monitor.value.columnUnit)
    })
    const maxValue = computed(() => {
      if (params.value.bounds.max === null) {
        return '+Inf'
      }
      return fmt(params.value.bounds.max, monitor.value.columnUnit)
    })
    const currentValue = computed(() => {
      return fmt(params.value.currentValue, monitor.value.columnUnit)
    })
    const currentValueVerbose = computed(() => {
      return numVerbose(params.value.currentValue)
    })
    const duration = computed(() => {
      const dur = params.value.numPointFiring * MINUTE
      const hours = Math.trunc(dur / HOUR)
      const minutes = Math.trunc((dur - hours * HOUR) / MINUTE)
      return formatDuration({ hours, minutes })
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

    const query = computed(() => {
      const tables = []
      for (let metric of monitor.value.metrics) {
        tables.push(`${metric.name} as $${metric.alias}`)
      }
      return joinQuery([monitor.value.query, `from ${tables.join(', ')}`])
    })

    const routeForSpans = computed(() => {
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

    const markPoint = computed(() => {
      return {
        name: 'outlier',
        value: props.alert.params.currentValue,
        unit: monitor.value.columnUnit,
        time: props.alert.time,
      }
    })

    return {
      AlertStatus,

      params,
      minValue,
      maxValue,
      currentValue,
      currentValueVerbose,
      duration,

      query,
      timeseries,
      styledTimeseries,
      routeForSpans,
      markPoint,
    }
  },
})
</script>

<style lang="scss" scoped></style>
