<template>
  <div>
    <h2 class="text-subtitle-1 font-weight-bold">{{ metric }}</h2>

    <MetricChart
      :loading="loading"
      :resolved="resolved"
      :timeseries="timeseries"
      :time="time"
      :event-bus="eventBus"
      :group="chartGroup"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Components
import MetricChart from '@/metrics/MetricChart.vue'

// Misc
import { EventBus } from '@/models/eventbus'
import { TimeseriesGroup } from '@/tracing/use-timeseries'

export default defineComponent({
  name: 'TimeseriesMetric',
  components: { MetricChart },

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    resolved: {
      type: Boolean,
      required: true,
    },
    metric: {
      type: String,
      required: true,
    },
    unit: {
      type: String,
      default: undefined,
    },
    groups: {
      type: Array as PropType<TimeseriesGroup[]>,
      required: true,
    },
    time: {
      type: Array as PropType<string[]>,
      required: true,
    },
    eventBus: {
      type: Object as PropType<EventBus>,
      default: undefined,
    },
    chartGroup: {
      type: [String, Symbol],
      default: () => Symbol(),
    },
  },

  setup(props) {
    const timeseries = computed(() => {
      return props.groups.map((group, i) => {
        return {
          id: group._id,
          name: group._name,
          metric: props.metric,
          value: group[props.metric],
          unit: props.unit,
          color: group._color,
          lineWidth: 2,
        }
      })
    })

    return { timeseries }
  },
})
</script>

<style lang="scss" scoped></style>
