<template>
  <v-simple-table>
    <thead v-if="loading">
      <tr class="v-data-table__progress">
        <th colspan="99" class="column">
          <v-progress-linear height="2" absolute indeterminate />
        </th>
      </tr>
    </thead>

    <tbody v-if="!results.length">
      <tr class="v-data-table__empty-wrapper">
        <td colspan="99">There are no any logs for the selected date range and filters.</td>
      </tr>
    </tbody>

    <tbody>
      <template v-for="(result, i) in results">
        <LogTableRow
          v-for="(value, j) in result.values"
          :key="`${i}-${j}`"
          :labels="result.stream"
          :timestamp="value[0]"
          :line="value[1]"
        />
      </template>
    </tbody>
  </v-simple-table>
</template>

<script lang="ts">
import { defineComponent, PropType } from '@vue/composition-api'

// Composables
import { Result } from '@/components/loki/logql'

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
    results: {
      type: Array as PropType<Result[]>,
      required: true,
    },
  },

  setup() {
    return {}
  },
})
</script>

<style lang="scss" scoped></style>
