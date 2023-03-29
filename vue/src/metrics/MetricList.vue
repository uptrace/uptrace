<template>
  <div>
    <v-row v-for="(metric, index) in activeMetrics" :key="metric.alias" dense align="center">
      <v-col cols="auto">
        <v-btn icon title="Remove metric" @click="removeMetric(index, metric)">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-col>
      <v-col cols="auto">
        ${{ metric.alias }} ({{ metric.name }})
        <v-chip
          v-if="metric.instrument"
          label
          color="grey lighten-4"
          title="Instrument"
          class="ml-2"
          >{{ metric.instrument }}</v-chip
        >
        <v-chip v-if="metric.unit" label color="grey lighten-4" title="Unit" class="ml-2">{{
          metric.unit
        }}</v-chip>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Utitlities
import { Metric, MetricAlias } from '@/metrics/types'

export default defineComponent({
  name: 'MetricList',

  props: {
    value: {
      type: Array as PropType<MetricAlias[]>,
      required: true,
    },
    metrics: {
      type: Array as PropType<Metric[]>,
      required: true,
    },
  },

  setup(props) {
    const activeMetrics = computed(() => {
      return props.value.map((metricAlias) => {
        const metric = props.metrics.find((m) => m.name === metricAlias.name)
        return {
          ...metric,
          ...metricAlias,
        }
      })
    })

    return { activeMetrics }
  },
})
</script>

<style lang="scss" scoped></style>
