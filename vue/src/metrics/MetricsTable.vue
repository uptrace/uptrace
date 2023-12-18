<template>
  <v-data-table
    :loading="loading"
    :headers="headers"
    :items="metrics"
    item-key="name"
    sort-by="numTimeseries"
    :sort-desc="true"
    must-sort
    :hide-default-footer="metrics.length <= 10"
    no-data-text="There are no metrics"
  >
    <template #item="{ item }">
      <tr class="cursor-pointer" @click="$emit('click:item', item)">
        <td>
          <div class="text-subtitle-2">
            <span>{{ item.name }}</span>
            <span class="ml-1 font-weight-regular text--secondary"> {{ item.description }}</span>
          </div>

          <div v-if="item.attrKeys" class="text-caption text--secondary">
            {{ item.attrKeys.join(', ') }}
          </div>
        </td>
        <td>
          {{ item.instrument }}
        </td>
        <td class="text-right">
          {{ item.numTimeseries }}
        </td>
      </tr>
    </template>
  </v-data-table>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Misc
import { ExploredMetric } from '@/metrics/use-metrics'

export default defineComponent({
  name: 'MetricsTable',

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    metrics: {
      type: Array as PropType<ExploredMetric[]>,
      required: true,
    },
  },

  setup() {
    const headers = computed(() => {
      const headers = []
      headers.push({ text: 'Metric', value: 'name', sortable: true, align: 'start' })
      headers.push({ text: 'Instrument', value: 'instrument', sortable: true, align: 'start' })
      headers.push({ text: 'Timeseries', value: 'numTimeseries', sortable: true, align: 'end' })
      return headers
    })

    return { headers }
  },
})
</script>

<style lang="scss" scoped>
td {
  height: 60px !important;
}
</style>
