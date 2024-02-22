<template>
  <div>
    <portal to="navigation">
      <v-tabs :key="$route.fullPath" background-color="transparent">
        <v-tab :to="{ name: 'DashboardList' }" exact-path>Dashboards</v-tab>
        <v-tab :to="{ name: 'MetricsExplore' }" exact-path>Explore Metrics</v-tab>
      </v-tabs>
    </portal>

    <HelpCard v-if="metrics.noData" :loading="metrics.loading" show-reload />
    <router-view v-else name="metrics" :date-range="dateRange" />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useMetrics } from '@/metrics/use-metrics'

// Components
import HelpCard from '@/metrics/HelpCard.vue'

export default defineComponent({
  name: 'MetricsLayout',
  components: { HelpCard },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
  },

  setup(props) {
    const metrics = useMetrics()

    return { metrics }
  },
})
</script>

<style lang="scss" scoped>
.border {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
  background-color: map-get($grey, 'lighten-5');
}
</style>
