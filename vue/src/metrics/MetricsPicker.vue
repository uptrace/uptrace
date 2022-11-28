<template>
  <div>
    <MetricPicker
      v-for="(metric, index) in activeMetrics"
      :key="index"
      :value="metric"
      :index="index"
      :metrics="metrics"
      :active-metrics="value"
      :uql="uql"
      :disabled="disabled"
      :required="index === 0"
      :class="{ 'mt-1': index > 0 }"
      @click:apply="applyMetric(metric, $event)"
      @click:remove="removeMetric"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, reactive, watch, PropType } from 'vue'

// Composables
import { UseUql } from '@/use/uql'
import { defaultMetricColumn } from '@/metrics/use-metrics'
import { hasMetricAlias } from '@/metrics/use-query'
import { Metric, MetricAlias } from '@/metrics/types'

// Components
import MetricPicker from '@/metrics/MetricPicker.vue'

export default defineComponent({
  name: 'MetricsPicker',
  components: { MetricPicker },

  props: {
    value: {
      type: Array as PropType<MetricAlias[]>,
      required: true,
    },
    metrics: {
      type: Array as PropType<Metric[]>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const activeMetrics = computed(() => {
      const metrics = props.value.slice()
      if (metrics.length < 6) {
        metrics.push({ name: '', alias: '' })
      }
      return metrics
    })

    const reactiveMetrics = computed(() => {
      return activeMetrics.value.map((metric) => reactive(metric))
    })

    const filledMetrics = computed(() => {
      return reactiveMetrics.value.filter((v) => v.name && v.alias)
    })

    watch(
      () => filledMetrics.value.length,
      () => {
        ctx.emit('input', filledMetrics.value)
      },
    )

    function applyMetric(metric: MetricAlias, newMetric: MetricAlias) {
      if (hasMetricAlias(props.uql.query, metric.alias)) {
        removeMetricAlias(metric.alias)
        addMetric(newMetric.name, newMetric.alias)
      } else {
        addMetric(newMetric.name, newMetric.alias)
      }

      metric.name = newMetric.name
      metric.alias = newMetric.alias
    }

    function removeMetricAlias(alias: string) {
      props.uql.parts = props.uql.parts.filter((part) => part.query.indexOf('$' + alias) === -1)
    }

    function addMetric(name: string, alias: string) {
      const metric = props.metrics.find((m) => m.name === name)
      if (metric) {
        const column = defaultMetricColumn(metric.instrument, alias)
        props.uql.query = props.uql.query + ' | ' + column
      }
    }

    function removeMetric(metric: MetricAlias) {
      props.uql.parts = props.uql.parts.filter((part) => {
        return !hasMetricAlias(part.query, metric.alias)
      })

      metric.name = ''
      metric.alias = ''
    }

    return {
      activeMetrics,

      applyMetric,
      removeMetric,
    }
  },
})
</script>

<style lang="scss" scoped></style>
