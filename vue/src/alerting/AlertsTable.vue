<template>
  <v-simple-table class="table-truncate-target">
    <colgroup>
      <col />
      <col class="target" />
      <col />
      <col />
    </colgroup>

    <thead v-show="loading">
      <tr class="v-data-table__progress">
        <th colspan="99" class="column">
          <v-progress-linear height="2" absolute indeterminate />
        </th>
      </tr>
    </thead>

    <tbody v-if="!alerts.length">
      <tr class="v-data-table__empty-wrapper">
        <td colspan="99" class="py-16 text-subtitle-1">
          <v-icon size="48" class="mb-4">mdi-magnify</v-icon>
          <p>You don't have alerts that match selected filters.</p>
        </td>
      </tr>
    </tbody>

    <tbody>
      <AlertsTableRow
        v-for="alert in alerts"
        :key="alert.id"
        :alert="alert"
        @click:alert="$emit('click:alert', alert)"
        @click:chip="$emit('click:chip', $event)"
      >
        <template #prepend-column>
          <slot name="prepend-column" :alert="alert" />
        </template>
      </AlertsTableRow>
    </tbody>
  </v-simple-table>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { UseOrder } from '@/use/order'
import { Alert } from '@/alerting/use-alerts'

// Components
import AlertsTableRow from '@/alerting/AlertsTableRow.vue'

export default defineComponent({
  name: 'AlertsTable',
  components: {
    AlertsTableRow,
  },

  props: {
    alerts: {
      type: Array as PropType<Alert[]>,
      required: true,
    },
    order: {
      type: Object as PropType<UseOrder>,
      required: true,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },

  setup() {
    return {}
  },
})
</script>

<style lang="scss" scoped>
td {
  height: 80px !important;
}

tr.no-border th {
  border-style: hidden !important;
}

.v-input--checkbox ::v-deep .v-input--selection-controls__input {
  margin-right: 0 !important;
}
</style>
