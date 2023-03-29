<template>
  <v-icon :title="type">{{ icon }}</v-icon>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { MonitorType } from '@/alerting/use-monitors'

export default defineComponent({
  name: 'MonitorTypeIcon',

  props: {
    type: {
      type: String as PropType<MonitorType>,
      required: true,
    },
  },

  setup(props) {
    const icon = computed(() => {
      switch (props.type) {
        case MonitorType.Metric:
          return 'mdi-chart-line'
        case MonitorType.Error:
          return 'mdi-bug-outline'
        default:
          return 'mdi-alert'
      }
    })

    return { icon }
  },
})
</script>

<style lang="scss" scoped></style>
