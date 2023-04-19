<template>
  <tr @click="$emit('click')">
    <slot :row-id="rowId" :metrics="metrics" :value="defaultValue" :time="timeseries.time" />
  </tr>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Compsables
import { AxiosParams } from '@/use/watch-axios'
import { useTimeseries } from '@/metrics/use-query'
import { Timeseries } from '@/metrics/types'

export default defineComponent({
  name: 'TimeseriesTableRow',

  props: {
    axiosParams: {
      type: Object as PropType<AxiosParams>,
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
          query += ' | ' + props.query
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

    const defaultValue = computed(() => {
      const value = timeseries.time.slice() as unknown as number[]
      value.fill(0)
      return value
    })

    return { rowId: Symbol(), timeseries, metrics, defaultValue }
  },
})
</script>

<style lang="scss" scoped></style>
