<template>
  <v-simple-table>
    <thead v-if="loading">
      <tr class="v-data-table__progress">
        <th colspan="99" class="column">
          <v-progress-linear height="2" absolute indeterminate />
        </th>
      </tr>
    </thead>

    <tbody v-if="!streams.length">
      <tr class="v-data-table__empty-wrapper">
        <td colspan="99">There are no any logs for the selected date range and filters.</td>
      </tr>
    </tbody>

    <tbody>
      <template v-for="(stream, i) in streams">
        <LogsTableRow
          v-for="(value, j) in stream.values"
          :key="`${i}-${j}`"
          :labels="stream.stream"
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
import { Stream } from '@/components/loki/logql'

// Components
import LogsTableRow from '@/components/loki/LogsTableRow.vue'

export default defineComponent({
  name: 'LogsTable',
  components: { LogsTableRow },

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    streams: {
      type: Array as PropType<Stream[]>,
      required: true,
    },
  },

  setup() {
    return {}
  },
})
</script>

<style lang="scss" scoped></style>
