<template>
  <SparklineChart
    :name="AttrKey.spanCountPerMin"
    :line="timeseries.data[AttrKey.spanCountPerMin] ?? []"
    :time="timeseries.time"
  />
</template>

<script lang="ts">
import { defineComponent, watch, PropType } from 'vue'

// Composables
import { useDateRange } from '@/use/date-range'
import { ErrorAlert } from '@/alerting/use-alerts'
import { useGroupTimeseries } from '@/tracing/use-explore-spans'

// Components
import SparklineChart from '@/components/SparklineChart.vue'

// Misc
import { AttrKey } from '@/models/otel'
import { HOUR } from '@/util/fmt/date'

export default defineComponent({
  name: 'AlertSparkline',
  components: { SparklineChart },

  props: {
    alert: {
      type: Object as PropType<ErrorAlert>,
      required: true,
    },
  },

  setup(props) {
    const dateRange = useDateRange()

    const timeseries = useGroupTimeseries(() => {
      return {
        ...dateRange.axiosParams(),
        query: `group by ${AttrKey.spanGroupId} | ${AttrKey.spanCountPerMin} | where ${AttrKey.spanGroupId} = ${props.alert.trackableId}`,
        column: AttrKey.spanCountPerMin,
      }
    })

    watch(
      () => props.alert.updatedAt,
      (updatedAt) => {
        dateRange.changeAround(updatedAt, HOUR)
      },
      { immediate: true },
    )

    return { AttrKey, timeseries }
  },
})
</script>

<style lang="scss" scoped></style>
