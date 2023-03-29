<template>
  <span class="text-no-wrap">
    <v-avatar size="12" v-bind="attrs" class="mr-2" />
    <span>{{ state }}</span>
  </span>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { NotifChannelState } from '@/alerting/use-notif-channels'

export default defineComponent({
  name: 'NotifChannelStateAvatar',

  props: {
    state: {
      type: String as PropType<NotifChannelState>,
      required: true,
    },
  },

  setup(props) {
    const attrs = computed(() => {
      switch (props.state) {
        case NotifChannelState.Delivering:
          return { color: 'success' }
        case NotifChannelState.Disabled:
          return { color: 'error', dark: true }
        default:
          return { color: 'grey' }
      }
    })

    return { attrs }
  },
})
</script>

<style lang="scss" scoped>
.v-avatar {
  height: 12px !important;
  min-width: 12px !important;
  width: 12px !important;
}
</style>
