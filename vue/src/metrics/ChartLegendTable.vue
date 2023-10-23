<template>
  <v-data-table
    ref="table"
    v-model="selectedItems"
    :loading="loading"
    :headers="headers"
    :items="timeseries"
    :items-per-page="9"
    :hide-default-footer="timeseries.length <= 9"
    no-data-text="There are no metrics"
    dense
    :show-expand="showExpand"
    single-expand
    @current-items="onCurrentItems($event)"
  >
    <template #item="{ item, isSelected, select, isExpanded, expand }">
      <tr
        :class="{ 'cursor-pointer': 'click' in $listeners, 'text--secondary': !isSelected }"
        @mouseenter="$emit('hover:item', { item: item, hover: true })"
        @mouseleave="$emit('hover:item', { item: item, hover: false })"
        @click="expand(!isExpanded)"
      >
        <td class="cursor-pointer text-caption" @click.stop="select(!isSelected)">
          <v-icon size="16" :color="isSelected ? item.color : 'grey'" class="mr-1"
            >mdi-circle</v-icon
          >
          <span>{{ truncateMiddle(item.name, maxLength) }}</span>
        </td>
        <td
          v-if="values.indexOf(LegendValue.Avg) >= 0"
          class="text-right text-caption font-weight-medium"
        >
          <NumValue :value="item.avg" :unit="item.unit" short />
        </td>
        <td
          v-if="values.indexOf(LegendValue.Last) >= 0"
          class="text-right text-caption font-weight-medium"
        >
          <NumValue :value="item.last" :unit="item.unit" short />
        </td>
        <td
          v-if="values.indexOf(LegendValue.Min) >= 0"
          class="text-right text-caption font-weight-medium"
        >
          <NumValue :value="item.min" :unit="item.unit" short />
        </td>
        <td
          v-if="values.indexOf(LegendValue.Max) >= 0"
          class="text-right text-caption font-weight-medium"
        >
          <NumValue :value="item.max" :unit="item.unit" short />
        </td>
        <td v-if="showExpand">
          <v-btn v-if="isExpanded" icon title="Hide spans" @click.stop="expand(false)">
            <v-icon>mdi-chevron-up</v-icon>
          </v-btn>
          <v-btn v-else icon title="View spans" @click.stop="expand(true)">
            <v-icon>mdi-chevron-down</v-icon>
          </v-btn>
        </td>
      </tr>
    </template>
    <template #expanded-item="{ headers, item }">
      <slot name="expanded-item" :headers="headers" :item="item" :expand-item="table.expand" />
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Utilities
import { StyledTimeseries, LegendValue } from '@/metrics/types'
import { truncateMiddle } from '@/util/string'

export default defineComponent({
  name: 'ChartLegendTable',

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
    maxLength: {
      type: Number,
      default: 40,
    },
  },

  setup(props, ctx) {
    const table = shallowRef()
    const selectedItems = shallowRef<StyledTimeseries[]>()

    const showExpand = computed(() => {
      return 'expanded-item' in ctx.slots
    })

    const headers = computed(() => {
      const headers = []

      headers.push({ text: 'metric', value: 'name', sortable: true, align: 'start' })

      for (let value of props.values) {
        switch (value) {
          case LegendValue.Avg:
            headers.push({ text: 'avg', value: 'avg', sortable: true, align: 'end' })
            break
          case LegendValue.Last:
            headers.push({ text: 'last', value: 'last', sortable: true, align: 'end' })
            break
          case LegendValue.Min:
            headers.push({ text: 'min', value: 'min', sortable: true, align: 'end' })
            break
          case LegendValue.Max:
            headers.push({ text: 'max', value: 'max', sortable: true, align: 'end' })
            break
          default:
            console.error(`unsupported legend value: ${value}`)
        }
      }

      if (showExpand.value) {
        headers.push({ text: '', value: 'data-table-expand', sortable: false })
      }

      return headers
    })

    function onCurrentItems(items: StyledTimeseries[]) {
      selectedItems.value = items
    }

    watch(selectedItems, (items) => {
      ctx.emit('current-items', items)
    })

    return {
      LegendValue,
      table,
      selectedItems,
      showExpand,
      headers,
      onCurrentItems,
      truncateMiddle,
    }
  },
})
</script>

<style lang="scss" scoped>
.v-data-table ::v-deep .v-data-table__wrapper > table {
  & > tbody > tr > td,
  & > tbody > tr > th,
  & > thead > tr > td,
  & > thead > tr > th,
  & > tfoot > tr > td,
  & > tfoot > tr > th {
    height: 28px !important;
    padding: 0 3px;
  }
}
</style>
