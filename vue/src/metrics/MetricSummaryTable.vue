<template>
  <v-data-table
    :headers="headers"
    :items="timeseries"
    :hide-default-footer="timeseries.length <= 10"
    no-data-text="There are no metrics"
    class="v-data-table--narrow"
  >
    <template #[`item.name`]="{ item }">
      <v-avatar size="12" :color="item.color" class="mr-2" />
      <span>{{ item.name }}</span>
    </template>
    <template #[`item.last`]="{ item }">
      <XNum :value="item.last" :unit="item.unit" short />
    </template>
    <template #[`item.avg`]="{ item }">
      <XNum :value="item.avg" :unit="item.unit" short />
    </template>
    <template #[`item.min`]="{ item }">
      <XNum :value="item.min" :unit="item.unit" short />
    </template>
    <template #[`item.max`]="{ item }">
      <XNum :value="item.max" :unit="item.unit" short />
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Utilities
import { Timeseries } from '@/metrics/use-query'

export default defineComponent({
  name: 'MetricSummaryTable',

  props: {
    timeseries: {
      type: Array as PropType<Timeseries[]>,
      required: true,
    },
  },

  setup() {
    const headers = computed(() => {
      const headers = []

      headers.push({ text: 'metric', value: 'name', sortable: true, align: 'start' })
      headers.push({ text: 'last', value: 'last', sortable: true, align: 'end' })
      headers.push({ text: 'avg', value: 'avg', sortable: true, align: 'end' })
      headers.push({ text: 'min', value: 'min', sortable: true, align: 'end' })
      headers.push({ text: 'max', value: 'max', sortable: true, align: 'end' })

      return headers
    })

    return { headers }
  },
})
</script>

<style lang="scss" scoped></style>
