<template>
  <div>
    <AlertCard :alert-id="$route.params.alertId" :fluid="$vuetify.breakpoint.mdAndDown" />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { useDateRange } from '@/use/date-range'

// Components
import AlertCard from '@/alerting/AlertCard.vue'

export default defineComponent({
  name: 'AlertShow',
  components: { AlertCard },

  setup() {
    const dateRange = useDateRange()

    return {
      dateRange,
      alert,
    }
  },
})
</script>

<style lang="scss" scoped></style>
