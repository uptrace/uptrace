<template>
  <v-data-table
    :loading="loading"
    :headers="headers"
    :items="items"
    hide-default-footer
    no-data-text="No gauges value"
    dense
    class="v-data-table--narrow"
  >
    <template #[`item.value`]="{ item }">
      <NumValue :value="item.value" :unit="item.unit" />
    </template>
    <template #[`item.unit`]="{ item }">
      <UnitPicker v-model="item.col.unit" target-class="py-0" />
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Components
import UnitPicker from '@/components/UnitPicker.vue'

// Misc
import { GaugeGridItem, MetricColumn, StyledColumnInfo } from '@/metrics/types'

export default defineComponent({
  name: 'GaugeValuesTable',
  components: { UnitPicker },

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    gridItem: {
      type: Object as PropType<GaugeGridItem>,
      required: true,
    },
    columns: {
      type: Array as PropType<StyledColumnInfo[]>,
      required: true,
    },
    values: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
    columnMap: {
      type: Object as PropType<Record<string, MetricColumn>>,
      required: true,
    },
  },

  setup(props) {
    const headers = [
      { text: 'Var name', value: 'name' },
      { text: 'Value', value: 'value' },
      { text: 'Unit', value: 'unit' },
    ]

    const items = computed(() => {
      const items = []
      for (let col of props.columns) {
        items.push({
          name: '${' + col.name + '}',
          value: props.values[col.name],
          unit: col.unit,
          col: props.columnMap[col.name],
        })
      }
      return items
    })

    return { headers, items }
  },
})
</script>

<style lang="scss" scoped></style>
