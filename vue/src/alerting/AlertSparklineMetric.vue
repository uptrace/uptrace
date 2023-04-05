<template>
  <SparklineChart
    v-if="sparkline"
    :name="monitor.column"
    :unit="monitor.columnUnit"
    :line="sparkline.value"
    :time="timeseries.time"
  />
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { useDateRange } from '@/use/date-range'
import { MetricAlert } from '@/alerting/use-alerts'
import { useTimeseries } from '@/metrics/use-query'

// Components
import SparklineChart from '@/components/SparklineChart.vue'

// Utilities
import { Timeseries } from '@/metrics/types'
import { AttrKey } from '@/models/otel'
import { HOUR } from '@/util/fmt/date'

export default defineComponent({
  name: 'AlertSparklineMetric',
  components: { SparklineChart },

  props: {
    alert: {
      type: Object as PropType<MetricAlert>,
      required: true,
    },
  },

  setup(props) {
    const dateRange = useDateRange()
    watch(
      () => props.alert.updatedAt,
      (updatedAt) => {
        dateRange.changeAround(updatedAt, HOUR)
      },
      { immediate: true },
    )

    const monitor = computed(() => {
      return props.alert.params.monitor
    })

    const timeseries = useTimeseries(() => {
      return {
        ...dateRange.axiosParams(),
        metric: monitor.value.metrics.map((m) => m.name),
        alias: monitor.value.metrics.map((m) => m.alias),
        query: monitor.value.query,
      }
    })

    const sparkline = computed((): Timeseries | undefined => {
      return timeseries.items.find((ts) => ts.metric === monitor.value.column)
    })

    return { AttrKey, monitor, timeseries, sparkline }
  },
})
</script>

<style lang="scss" scoped></style>
