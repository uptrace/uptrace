<template>
  <span>
    <v-btn
      v-if="dashboard.pinned"
      :loading="dashMan.pending"
      icon
      title="Unpin dashboard"
      @click="unpinDashboard"
    >
      <v-icon color="green darken-2">mdi-pin</v-icon>
    </v-btn>
    <v-btn v-else :loading="dashMan.pending" icon title="Pin dashboard" @click="pinDashboard">
      <v-icon>mdi-pin-outline</v-icon>
    </v-btn>
  </span>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useDashboardManager } from '@/metrics/use-dashboards'

// Misc
import { Dashboard } from '@/metrics/types'

export default defineComponent({
  name: 'DashPinBtn',

  props: {
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
  },

  setup(props, ctx) {
    const dashMan = useDashboardManager()

    function pinDashboard() {
      dashMan.pin(props.dashboard).then((dash) => {
        ctx.emit('update', dash)
      })
    }

    function unpinDashboard() {
      dashMan.unpin(props.dashboard).then((dash) => {
        ctx.emit('update', dash)
      })
    }

    return { dashMan, pinDashboard, unpinDashboard }
  },
})
</script>

<style lang="scss" scoped></style>
