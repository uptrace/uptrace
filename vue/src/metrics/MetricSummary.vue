<template>
  <v-expansion-panels flat>
    <v-expansion-panel>
      <v-expansion-panel-header color="grey lighten-4">
        Summary ({{ timeseries.length }})
      </v-expansion-panel-header>
      <v-expansion-panel-content>
        <MetricSummaryTable :timeseries="timeseries" />
      </v-expansion-panel-content>
    </v-expansion-panel>
  </v-expansion-panels>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Components
import MetricSummaryTable from '@/metrics/MetricSummaryTable.vue'

// Utilities
import { Timeseries } from '@/metrics/types'

export default defineComponent({
  name: 'MetricSummary',
  components: { MetricSummaryTable },

  props: {
    timeseries: {
      type: Array as PropType<Timeseries[]>,
      required: true,
    },
  },
})
</script>

<style lang="scss" scoped>
.v-expansion-panel ::v-deep .v-expansion-panel-content__wrap {
  padding-left: 4px !important;
  padding-right: 4px !important;
}
</style>
