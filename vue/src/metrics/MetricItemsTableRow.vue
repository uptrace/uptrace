<template>
  <tr @click="$emit('click')">
    <slot :metrics="metrics" :default-timeseries="defaultTimeseries" />
  </tr>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Compsables
import { useRouter } from '@/use/router'
import { useTimeseries } from '@/metrics/use-query'

export default defineComponent({
  name: 'MetricItemsTableRow',

  props: {
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
    query: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const timeseries = useTimeseries(() => {
      let baseQuery = props.axiosParams.base_query ?? ''
      if (props.query) {
        baseQuery += ' | ' + props.query
      }

      const { projectId } = route.value.params
      return {
        url: `/api/v1/metrics/${projectId}/timeseries`,
        params: {
          ...props.axiosParams,
          base_query: baseQuery,
        },
        cache: true,
      }
    })

    const metrics = computed(() => {
      const metrics = {}
      for (let ts of timeseries.items) {
        metrics[ts.metric] = ts
      }
      return metrics
    })

    const defaultTimeseries = computed(() => {
      if (!timeseries.items.length) {
        return { value: [], time: [] }
      }

      const ts = timeseries.items[0]
      return { value: ts.value.slice().fill(0), time: ts.time }
    })

    return { metrics, defaultTimeseries }
  },
})
</script>

<style lang="scss" scoped></style>
