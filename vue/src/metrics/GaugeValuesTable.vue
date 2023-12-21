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
  </v-data-table>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Misc
import { StyledColumnInfo } from '@/metrics/types'

export default defineComponent({
  name: 'GaugeValuesTable',

  props: {
    loading: {
      type: Boolean,
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
  },

  setup(props) {
    const headers = [
      { text: 'Var name', value: 'name' },
      { text: 'Value', value: 'value' },
    ]

    const items = computed(() => {
      const items = []
      for (let col of props.columns) {
        items.push({
          name: '${' + col.name + '}',
          value: props.values[col.name],
        })
      }
      return items
    })

    return { headers, items }
  },
})
</script>

<style lang="scss" scoped></style>
