<template>
  <div class="d-flex flex-wrap justify-center text-caption" :class="`flex-${direction}`">
    <div
      v-for="item in timeseries"
      :key="item.name"
      class="mx-2 d-flex align-center cursor-pointer"
      :class="{ 'text--secondary': !isSelected(item) }"
      @mouseenter="$emit('hover:item', { item: item, hover: true })"
      @mouseleave="$emit('hover:item', { item: item, hover: false })"
      @click.stop="toggle(item)"
    >
      <v-icon size="16" :color="isSelected(item) ? item.color : 'grey'" class="mr-1"
        >mdi-circle</v-icon
      >
      <div>{{ truncateMiddle(item.name, 60) }}</div>
      <div v-for="metric in values" :key="metric" class="ml-1 font-weight-medium">
        <NumValue
          :value="item[metric]"
          :unit="item.unit"
          format="short"
          :title="`${metric}: {0}`"
        />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType } from 'vue'

// Misc
import { StyledTimeseries, LegendValue } from '@/metrics/types'
import { truncateMiddle } from '@/util/string'

export default defineComponent({
  name: 'ChartLegendList',

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    timeseries: {
      type: Array as PropType<StyledTimeseries[]>,
      required: true,
    },
    values: {
      type: Array as PropType<LegendValue[]>,
      default: () => [LegendValue.Avg, LegendValue.Last, LegendValue.Min, LegendValue.Max],
    },
    direction: {
      type: String,
      default: 'row',
    },
  },

  setup(props, ctx) {
    const selectedTimeseries = shallowRef<StyledTimeseries[]>([])
    watch(
      () => props.timeseries,
      (timeseries) => {
        selectedTimeseries.value = timeseries
      },
      { immediate: true },
    )
    watch(
      selectedTimeseries,
      (selectedTimeseries) => {
        ctx.emit('current-items', selectedTimeseries)
      },
      { immediate: true },
    )

    function toggle(ts: StyledTimeseries) {
      const items = selectedTimeseries.value.slice()
      const index = items.findIndex((item) => item.id === ts.id)
      if (index >= 0) {
        items.splice(index, 1)
      } else {
        items.push(ts)
      }
      selectedTimeseries.value = items
    }

    function isSelected(ts: StyledTimeseries): boolean {
      const index = selectedTimeseries.value.findIndex((item) => item.id === ts.id)
      return index >= 0
    }

    return {
      selectedTimeseries,
      toggle,
      isSelected,

      truncateMiddle,
    }
  },
})
</script>

<style lang="scss" scoped></style>
