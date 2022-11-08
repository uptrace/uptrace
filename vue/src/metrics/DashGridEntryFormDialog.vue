<template>
  <v-dialog :value="value" max-width="1200" @input="$emit('input', $event)">
    <DashGridEntryForm
      :date-range="dateRange"
      :metrics="metrics"
      :dashboard="dashboard"
      :dash-entry="dashEntry"
      :timeseries="timeseries"
      @click:save="onSave()"
      @click:cancel="onCancel()"
    />
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { UseMetrics } from '@/metrics/use-metrics'
import { UseDashboard, DashEntry } from '@/metrics/use-dashboards'
import { UseTimeseries } from '@/metrics/use-query'

// Components
import DashGridEntryForm from '@/metrics/DashGridEntryForm.vue'

export default defineComponent({
  name: 'DashGridEntryFormDialog',
  components: {
    DashGridEntryForm,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },

    value: {
      type: Boolean,
      required: true,
    },
    metrics: {
      type: Object as PropType<UseMetrics>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<UseDashboard>,
      required: true,
    },
    dashEntry: {
      type: Object as PropType<DashEntry>,
      required: true,
    },
    timeseries: {
      type: Object as PropType<UseTimeseries>,
      required: true,
    },
  },

  setup(props, ctx) {
    function onSave() {
      ctx.emit('change')
      ctx.emit('input', false)
    }

    function onCancel() {
      ctx.emit('change')
      ctx.emit('input', false)
    }

    return { onSave, onCancel }
  },
})
</script>

<style lang="scss" scoped></style>
