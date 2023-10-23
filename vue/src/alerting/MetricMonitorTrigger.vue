<template>
  <span>
    <NumValue :value="alert.params.outlier" :unit="unit" />
    <template v-if="alert.params.firing === 1">
      <span class="mx-1">more than</span>
      <NumValue :value="alert.params.bounds.max" :unit="unit" />
    </template>
    <template v-else-if="alert.params.firing === -1">
      <span class="mx-1">less than</span>
      <NumValue :value="alert.params.bounds.min" :unit="unit" />
    </template>
  </span>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { MetricAlert } from '@/alerting/use-alerts'

export default defineComponent({
  name: 'MetricMonitorTrigger',

  props: {
    alert: {
      type: Object as PropType<MetricAlert>,
      required: true,
    },
    verbose: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const unit = computed(() => {
      return props.alert.params.monitor.columnUnit
    })

    return { unit }
  },
})
</script>

<style lang="scss" scoped></style>
