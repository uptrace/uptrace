<template>
  <v-simple-table>
    <thead v-if="loading">
      <tr class="v-data-table__progress">
        <th colspan="99" class="column">
          <v-progress-linear height="2" absolute indeterminate />
        </th>
      </tr>
    </thead>

    <tbody v-if="!logs.length">
      <tr class="v-data-table__empty-wrapper">
        <td colspan="99">There are no any logs for the selected date range and filters.</td>
      </tr>
    </tbody>

    <tbody>
      <LogTableRow
        v-for="(log, i) in logs"
        :key="i"
        :labels="labels"
        :timestamp="log[0]"
        :line="log[1]"
      />
    </tbody>
  </v-simple-table>
</template>

<script lang="ts">
import { defineComponent, PropType } from '@vue/composition-api'

// Composables
import { LogValue } from '@/components/loki/logql'

// Components
import LogTableRow from '@/components/loki/LogTableRow.vue'

export default defineComponent({
  name: 'LogsTable',
  components: { LogTableRow },

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    labels: {
      type: Object as PropType<Record<string, string>>,
      required: true,
    },
    logs: {
      type: Array as PropType<LogValue[]>,
      required: true,
    },
  },

  setup() {
    return {}
  },
})
</script>

<style lang="scss" scoped></style>
