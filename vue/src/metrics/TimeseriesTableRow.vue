<template>
  <tr @click="$emit('click')">
    <slot
      :row-id="rowId"
      :metrics="metrics"
      :empty-value="timeseries.emptyValue"
      :time="timeseries.time"
    />
  </tr>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Compsables
import { joinQuery } from '@/use/uql'
import { useTimeseries } from '@/metrics/use-query'

// Misc
import { Timeseries } from '@/metrics/types'

export default defineComponent({
  name: 'TimeseriesTableRow',

  props: {
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      default: undefined,
    },
    query: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const timeseries = useTimeseries(
      () => {
        if (!props.axiosParams) {
          return props.axiosParams
        }

        let query = props.axiosParams.query
        if (props.query) {
          query = joinQuery([props.axiosParams.query, props.query])
        }

        return {
          ...props.axiosParams,
          query,
        }
      },
      { cache: true },
    )

    const metrics = computed(() => {
      const metrics: Record<string, Timeseries> = {}
      for (let ts of timeseries.items) {
        metrics[ts.metric] = ts
      }
      return metrics
    })

    return { rowId: Symbol(), timeseries, metrics }
  },
})
</script>

<style lang="scss" scoped></style>
