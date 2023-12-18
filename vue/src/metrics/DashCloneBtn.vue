<template>
  <v-btn :loading="dashMan.pending" depressed small class="ml-2" @click="cloneDashboard"
    >Clone</v-btn
  >
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Misc
import { useDashboardManager, UseDashboard } from '@/metrics/use-dashboards'

export default defineComponent({
  name: 'DashCloneBtn',

  props: {
    dashboard: {
      type: Object as PropType<UseDashboard>,
      required: true,
    },
  },

  setup(props, ctx) {
    const dashMan = useDashboardManager()

    function cloneDashboard() {
      dashMan.clone(props.dashboard.data!).then((dash) => {
        ctx.emit('click:clone', dash)
      })
    }

    return { dashMan, cloneDashboard }
  },
})
</script>

<style lang="scss" scoped></style>
