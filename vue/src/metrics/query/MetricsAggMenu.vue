<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text :disabled="disabled" class="v-btn--filter" v-bind="attrs" v-on="on">
        Aggregate
      </v-btn>
    </template>

    <v-card>
      <v-list dense>
        <template v-for="item in items">
          <v-menu :key="item.metric.alias" open-on-hover offset-x transition="slide-x-transition">
            <template #activator="{ on, attrs }">
              <v-list-item v-bind="attrs" v-on="on">
                <v-list-item-content>
                  <v-list-item-title>${{ item.metric.alias }}</v-list-item-title>
                </v-list-item-content>
                <v-list-item-icon class="align-self-center">
                  <v-icon>mdi-menu-right</v-icon>
                </v-list-item-icon>
              </v-list-item>
            </template>

            <v-list dense>
              <v-list-item v-for="col in item.columns" :key="col.value" @click="aggBy(col.value)">
                <v-list-item-content>
                  <v-list-item-title>{{ col.value }}</v-list-item-title>
                  <v-list-item-subtitle v-if="col.hint">
                    {{ col.hint }}
                  </v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-menu>
        </template>
      </v-list>
    </v-card>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { UseUql } from '@/use/uql'

// Types
import { ActiveMetric as Metric, Instrument } from '@/metrics/types'

interface MetricItem {
  metric: Metric
  columns: ColumnItem[]
}

interface ColumnItem {
  value: string
  hint?: string
}

export default defineComponent({
  name: 'MetricsAggMenu',

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    metrics: {
      type: Array as PropType<Metric[]>,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const menu = shallowRef(false)

    const items = computed((): MetricItem[] => {
      const items = []
      for (let metric of props.metrics) {
        items.push({
          metric,
          columns: metricColumns(metric),
        })
      }
      return items
    })

    function aggBy(column: string) {
      const editor = props.uql.createEditor()
      editor.add(column)
      props.uql.commitEdits(editor)

      menu.value = false
    }

    return {
      menu,
      items,

      aggBy,
    }
  },
})

function metricColumns(metric: Metric): ColumnItem[] {
  const alias = '$' + metric.alias
  switch (metric.instrument) {
    case Instrument.Deleted:
      return []
    case Instrument.Gauge:
      return [
        { value: alias, hint: `same as avg(${alias})` },
        { value: `avg(${alias})`, hint: 'avg value, e.g. on different hostnames' },
        { value: `min(${alias})`, hint: 'minimim value' },
        { value: `max(${alias})`, hint: 'maximum value' },
        { value: `sum(${alias})`, hint: 'sum of values (usually invalid)' },
      ]
    case Instrument.Additive:
      return [
        { value: alias, hint: `same as sum(${alias})` },
        { value: `sum(${alias})`, hint: 'sum of values, e.g. on different hostnames' },
        { value: `avg(${alias})`, hint: 'avg value in different timeseries' },
        { value: `min(${alias})` },
        { value: `max(${alias})` },
      ]
    case Instrument.Counter:
      return [
        { value: `${alias}`, hint: 'sum of values' },
        { value: `per_min(${alias})`, hint: 'sum(value) / number_of_minutes' },
        { value: `per_sec(${alias})`, hint: 'sum(value) / number_of_seconds' },
      ]
    case Instrument.Histogram:
      return [
        { value: `p50(${alias})` },
        { value: `p75(${alias})` },
        { value: `p90(${alias})` },
        { value: `p95(${alias})` },
        { value: `p99(${alias})` },
        { value: `count(${alias})`, hint: 'number of occurrences' },
        { value: `per_min(${alias})`, hint: 'count() / number_of_minutes' },
        { value: `per_sec(${alias})`, hint: 'count() / number_of_seconds' },
        { value: `avg(${alias})` },
        { value: `min(${alias})` },
        { value: `max(${alias})` },
      ]
    default:
      throw new Error(`unknown instrument: ${metric.instrument}`)
  }
}
</script>

<style lang="scss" scoped>
.v-select.fit {
  min-width: min-content !important;
}

.v-select.fit .v-select__selection--comma {
  text-overflow: unset;
}

.no-transform ::v-deep .v-btn {
  padding: 0 12px !important;
  text-transform: none;
}
</style>
