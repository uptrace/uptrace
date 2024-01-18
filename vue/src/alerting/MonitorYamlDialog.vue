<template>
  <DialogCard :value="value" max-width="800" :title="monitor.name" @input="$emit('input', $event)">
    <div class="container--fixed-sm">
      <v-container>
        <PrismCode v-if="yaml.data" :code="yaml.data" language="yaml" />
        <v-skeleton-loader v-else type="article" />
      </v-container>
    </div>
  </DialogCard>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useYamlMonitor } from '@/alerting/use-monitors'

// Components
import DialogCard from '@/components/DialogCard.vue'

// Misc
import { Monitor } from '@/alerting/types'

export default defineComponent({
  name: 'MonitorYamlDialog',

  components: { DialogCard },

  props: {
    value: {
      type: Boolean,
      required: true,
    },
    monitor: {
      type: Object as PropType<Monitor>,
      required: true,
    },
  },

  setup(props) {
    const yaml = useYamlMonitor(props.monitor.id)

    return {
      yaml,
    }
  },
})
</script>

<style lang="scss" scoped></style>
