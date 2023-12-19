<template>
  <div>
    <PercentilesChart
      v-if="!percentiles.status.hasData() || 'p50' in percentiles.stats"
      :loading="percentiles.loading"
      :time="percentiles.stats.time"
      :count-per-min="percentiles.stats.rate"
      :errors-per-min="percentiles.stats.errorRate"
      :p50="percentiles.stats.p50"
      :p90="percentiles.stats.p90"
      :p99="percentiles.stats.p99"
      :max="percentiles.stats.max"
      :mark-point="markPoint"
      :annotations="annotations"
    />
    <EventRateChart
      v-else
      :loading="percentiles.loading"
      :time="percentiles.stats.time"
      :count-per-min="percentiles.stats.rate"
      :mark-point="markPoint"
      :annotations="annotations"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { usePercentiles } from '@/tracing/use-percentiles'

// Components
import PercentilesChart from '@/components/PercentilesChart.vue'
import EventRateChart from '@/components/EventRateChart.vue'
import { MarkPoint } from '@/util/chart'

// Misc
import { Annotation } from '@/org/types'

export default defineComponent({
  name: 'PercentilesChartLazy',
  components: { PercentilesChart, EventRateChart },

  props: {
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
    markPoint: {
      type: Object as PropType<MarkPoint>,
      default: undefined,
    },
    annotations: {
      type: Array as PropType<Annotation[]>,
      default: () => [],
    },
  },

  setup(props) {
    const route = useRoute()

    const percentiles = usePercentiles(() => {
      const { projectId } = route.value.params
      return {
        url: `/internal/v1/tracing/${projectId}/percentiles`,
        params: props.axiosParams,
        cache: true,
      }
    })

    return { percentiles }
  },
})
</script>

<style lang="scss" scoped></style>
