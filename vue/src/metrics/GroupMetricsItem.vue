<template>
  <GridColumnChart
    :loading="timeseries.loading"
    :resolved="timeseries.status.isResolved()"
    :timeseries="styledTimeseries"
    :time="timeseries.time"
    :legend="legend"
  />
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { defaultMetricAlias, defaultMetricQuery } from '@/metrics/use-metrics'
import { useTimeseries, useStyledTimeseries } from '@/metrics/use-query'

// Components
import GridColumnChart from '@/metrics/GridColumnChart.vue'

// Utilities
import { Metric, LegendType, LegendPlacement, LegendValue } from '@/metrics/types'

export default defineComponent({
  name: 'GroupMetricsItem',
  components: { GridColumnChart },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    metric: {
      type: Object as PropType<Metric>,
      required: true,
    },
    where: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const legend = computed(() => {
      return {
        type: LegendType.List,
        placement: LegendPlacement.Bottom,
        values: [LegendValue.Avg],
      }
    })

    const timeseries = useTimeseries(() => {
      const alias = defaultMetricAlias(props.metric.name)

      return {
        ...props.dateRange.axiosParams(),
        metric: props.metric.name,
        alias,
        query: defaultMetricQuery(props.metric.instrument, alias) + ' | ' + props.where,
      }
    })

    const styledTimeseries = useStyledTimeseries(
      computed(() => timeseries.items),
      computed(() => ({})),
      computed(() => ({})),
    )

    return { legend, timeseries, styledTimeseries }
  },
})
</script>

<style lang="scss" scoped></style>
