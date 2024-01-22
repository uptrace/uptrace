<template>
  <GaugeCard
    :loading="gaugeQuery.loading"
    :grid-item="gridItem"
    :columns="gaugeQuery.styledColumns"
    :values="gaugeQuery.values"
    :column-map="gridItem.params.columnMap"
    show-edit
    @click:edit="$emit('click:edit', $event)"
    @change="$emit('change', $event)"
  />
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { joinQuery, injectQueryStore } from '@/use/uql'
import { useGaugeQuery } from '@/metrics/use-gauges'

// Components
import GaugeCard from '@/metrics/GaugeCard.vue'

// Misc
import { GaugeGridItem } from '@/metrics/types'

export default defineComponent({
  name: 'GridItemGauge',
  components: { GaugeCard },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    gridItem: {
      type: Object as PropType<GaugeGridItem>,
      required: true,
    },
  },

  setup(props, ctx) {
    const { where } = injectQueryStore()
    const gaugeQuery = useGaugeQuery(
      () => {
        if (!props.gridItem.params.metrics.length || !props.gridItem.params.query) {
          return { _: undefined }
        }

        return {
          ...props.dateRange.axiosParams(),
          metric: props.gridItem.params.metrics.map((m) => m.name),
          alias: props.gridItem.params.metrics.map((m) => m.alias),
          query: joinQuery([props.gridItem.params.query, where.value]),
        }
      },
      computed(() => props.gridItem.params.columnMap),
    )

    watch(
      () => gaugeQuery.status,
      (status) => {
        if (status.isResolved()) {
          ctx.emit('ready')
        }
      },
    )
    watch(
      () => gaugeQuery.error,
      (error) => {
        ctx.emit('error', error)
      },
    )

    return { gaugeQuery }
  },
})
</script>

<style lang="scss" scoped></style>
