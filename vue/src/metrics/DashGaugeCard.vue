<template>
  <v-card outlined rounded="lg" class="py-4 px-5 border-bottom text-center">
    <v-tooltip top>
      <template #activator="{ on, attrs }">
        <div class="body-2 text-truncate blue-grey--text text--lighten-1" v-bind="attrs" v-on="on">
          {{ gauge.name }}
        </div>
      </template>
      <span>{{ gauge.description || gauge.name }}</span>
    </v-tooltip>

    <div class="pt-4 text-h5 text-truncate">
      {{ text }}
    </div>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { DashGauge } from '@/metrics/use-dashboards'
import { useGaugeQuery } from '@/metrics/use-query'

// Utilities
import { fmt } from '@/util/fmt'

export default defineComponent({
  name: 'DashGaugeCard',

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    gauge: {
      type: Object as PropType<DashGauge>,
      required: true,
    },
    baseQuery: {
      type: String,
      default: '',
    },
  },

  setup(props) {
    const gaugeQuery = useGaugeQuery(() => {
      if (!props.gauge.metrics.length) {
        return { _: undefined }
      }

      return {
        ...props.dateRange.axiosParams(),
        metrics: props.gauge.metrics.map((m) => m.name),
        aliases: props.gauge.metrics.map((m) => m.alias),
        query: props.gauge.query,
        base_query: props.baseQuery,
      }
    })

    const text = computed(() => {
      if (!gaugeQuery.columns.length) {
        return '-'
      }

      let text = props.gauge.template
      if (text) {
        for (let col of gaugeQuery.columns) {
          const val = gaugeQuery.values[col.name]
          if (val === undefined) {
            text = text.replaceAll('$' + col.name, '-')
            continue
          }
          const unit = props.gauge.columnMap[col.name]?.unit ?? col.unit
          text = text.replaceAll('$' + col.name, fmt(val, unit))
        }
        return text
      }

      const col = gaugeQuery.columns[0]
      const val = gaugeQuery.values[col.name]
      if (val === undefined) {
        return '-'
      }
      const unit = props.gauge.columnMap[col.name]?.unit ?? col.unit
      return fmt(val, unit)
    })

    return { text }
  },
})
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom: 6px map-get($blue, 'darken-2') solid;
}
</style>
