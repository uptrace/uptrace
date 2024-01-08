<template>
  <v-dialog :value="value" fullscreen @input="$emit('input', $event)">
    <DashTableForm
      v-if="value"
      :date-range="internalDateRange"
      :dashboard="dashboard"
      @saved="
        $emit('input', false)
        $emit('saved')
      "
      @click:cancel="$emit('input', false)"
    >
    </DashTableForm>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useDateRangeFrom, UseDateRange } from '@/use/date-range'

// Components
import DashTableForm from '@/metrics/DashTableForm.vue'

// Misc
import { Dashboard } from '@/metrics/types'

export default defineComponent({
  name: 'DashTableFormDialog',
  components: { DashTableForm },

  props: {
    value: {
      type: Boolean,
      required: true,
    },
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
  },

  setup(props, ctx) {
    const internalDateRange = useDateRangeFrom(props.dateRange)
    return { internalDateRange }
  },
})
</script>

<style lang="scss" scoped></style>
