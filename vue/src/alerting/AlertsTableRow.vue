<template>
  <tr class="cursor-pointer" @click="$emit('click:alert', alert)">
    <slot name="prepend-column" :alert="alert" />
    <td class="target">
      <div class="mb-1 font-weight-medium">
        <v-icon :color="statusColor" class="mr-1">{{
          alert.status === AlertStatus.Open
            ? 'mdi-alert-circle-outline'
            : 'mdi-alert-circle-check-outline'
        }}</v-icon>
        {{ alert.name }}
      </div>

      <AlertChips :alert="alert" @click:chip="$emit('click:chip', $event)" />
      <span class="ml-3 text-caption text--secondary">
        <span>Created </span>
        <DateValue :value="alert.createdAt" format="relative" />
      </span>
      <span v-if="alert.updatedAt !== alert.createdAt" class="ml-3 text-caption text--secondary">
        <span>Updated </span>
        <DateValue :value="alert.updatedAt" format="relative" />
      </span>
    </td>
    <td class="text-center text-caption font-weight-medium">
      <template v-if="alert.type === AlertType.Metric">
        <MetricMonitorTrigger :alert="alert" />
        <AlertSparklineMetric :alert="alert" />
      </template>
      <template v-else>
        <div v-if="alert.params.spanCount">
          <NumValue :value="alert.params.spanCount" /> occurrences
        </div>
        <AlertSparklineError :alert="alert" />
      </template>
    </td>
  </tr>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Compsables
import { Alert, AlertType, AlertStatus } from '@/alerting/use-alerts'

// Components
import MetricMonitorTrigger from '@/alerting/MetricMonitorTrigger.vue'
import AlertSparklineError from '@/alerting/AlertSparklineError.vue'
import AlertSparklineMetric from '@/alerting/AlertSparklineMetric.vue'
import AlertChips from '@/alerting/AlertChips.vue'

export default defineComponent({
  name: 'AlertsTableRow',
  components: { MetricMonitorTrigger, AlertSparklineError, AlertSparklineMetric, AlertChips },

  props: {
    alert: {
      type: Object as PropType<Alert>,
      required: true,
    },
  },

  setup(props) {
    const statusColor = computed(() => {
      switch (props.alert.status) {
        case AlertStatus.Open:
          return 'red darken-2'
        default:
          return 'green darken-2'
      }
    })

    return {
      AlertStatus,
      AlertType,
      statusColor,
    }
  },
})
</script>

<style lang="scss" scoped>
td {
  height: 80px !important;
}
</style>
