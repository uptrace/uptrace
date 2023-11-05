<template>
  <PagedGroupsCard
    :date-range="dateRange"
    :systems="systems"
    :loading="groups.loading"
    :groups="groups.items"
    :columns="groups.columns"
    :plottable-columns="groups.plottableColumns"
    :plotted-columns="plottedColumns"
    :show-plotted-column-items="showPlottedColumnItems"
    :hide-actions="hideActions"
    :order="groups.order"
    :axios-params="groups.axiosParams"
    @update:plotted-columns="$emit('@update:plotted-columns', $event)"
    @update:num-group="$emit('@update:num-group', $event)"
  />
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useGroups } from '@/tracing/use-explore-spans'

export default defineComponent({
  name: 'PagedGroupsCardLazy',

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Array as PropType<string[]>,
      required: true,
    },
    plottedColumns: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    showPlottedColumnItems: {
      type: Boolean,
      default: false,
    },
    hideActions: {
      type: Boolean,
      default: false,
    },
    query: {
      type: String,
      required: true,
    },
  },

  setup(props, ctx) {
    const groups = useGroups(() => {
      return {
        ...props.dateRange.axiosParams(),
        system: props.systems,
        query: props.query,
      }
    })

    return {
      groups,
    }
  },
})
</script>

<style lang="scss" scoped></style>
