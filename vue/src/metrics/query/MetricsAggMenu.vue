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

// Misc
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

    case Instrument.Counter:
      return [
        { value: alias, hint: `sum of timeseries; sum($result) in table` },
        { value: `per_min(${alias})`, hint: 'sum(value) / _minutes; avg value in table' },
        { value: `per_sec(${alias})`, hint: 'sum(value) / _seconds; avg value in table' },
        { value: `last(${alias})`, hint: `sum of timeseries; last value in table` },
      ]

    case Instrument.Gauge:
      return [
        { value: alias, hint: `avg of timeseries; last value in table` },
        { value: `avg(${alias})`, hint: 'avg of timeseries; avg value in table' },
        { value: `min(${alias})`, hint: 'min of timeseries; min value in table' },
        { value: `max(${alias})`, hint: 'max of timeseries; max value in table' },
        { value: `sum(${alias})`, hint: 'sum of timeseries; sum in table (compat)' },
        { value: `last(sum(${alias}))`, hint: 'sum of timeseries; last value in table (compat)' },
        { value: `per_min(${alias})`, hint: '$metric / _minutes; avg in table (compat)' },
        { value: `per_sec(${alias})`, hint: '$metric / _seconds; avg in table (compat)' },
        {
          value: `delta(${alias})`,
          hint: 'diff between curr and prev values; sum in table (compat)',
        },
      ]

    case Instrument.Additive:
      return [
        { value: alias, hint: `sum of timeseries; last value in table` },
        { value: `sum(${alias})`, hint: 'sum of timeseries; sum in table' },
        { value: `avg(${alias})`, hint: 'avg of timeseries; avg in table' },
        { value: `last(avg(${alias}))`, hint: 'avg of timeseries; last value in table' },
        { value: `min(${alias})`, hint: 'min of timeseries; min in table' },
        { value: `max(${alias})`, hint: 'max of timeseries; max in table' },
        { value: `per_min($metric)`, hint: '$metric / _minutes; avg in table' },
        { value: `per_sec($metric)`, hint: '$metric / _seconds; avg in table' },
        {
          value: `delta($metric)`,
          hint: 'diff between curr and prev values; sum in table (compat)',
        },
      ]

    case Instrument.Histogram:
      return [
        { value: `count(${alias})`, hint: 'number of observed values; sum in table' },
        { value: `per_min(count(${alias}))`, hint: 'count() / _minutes; avg in table' },
        { value: `per_sec(count(${alias}))`, hint: 'count() / _seconds; avg in table' },
        { value: `p50(${alias})`, hint: 'p50 of timeseries; last value in table' },
        { value: `p75(${alias})`, hint: 'p75 of timeseries; last value in table' },
        { value: `p90(${alias})`, hint: 'p90 of timeseries; last value in table' },
        { value: `p95(${alias})`, hint: 'p95 of timeseries; last value in table' },
        { value: `p99(${alias})`, hint: 'p99 of timeseries; last value in table' },
        { value: `avg(${alias})`, hint: 'sum($metric) / count($metric); avg in table' },
        {
          value: `last(avg(${alias}))`,
          hint: 'sum($metric) / count($metric); last value in table',
        },
        { value: `min(${alias})`, hint: 'min of timeseries; min in table' },
        { value: `max(${alias})`, hint: 'max of timeseries; max in table' },
      ]

    case Instrument.Summary:
      return [
        { value: `avg(${alias})`, hint: 'avg of timeseries; avg in table' },
        { value: `last(avg(${alias}))`, hint: 'avg of timeseries; last value in table' },
        { value: `min(${alias})`, hint: 'min of timeseries; min in table' },
        { value: `max(${alias})`, hint: 'max of timeseries; max in table' },
        { value: `count(${alias})`, hint: 'number of observed values; sum in table' },
        { value: `sum(${alias})`, hint: 'sum of timeseries; sum in table' },
        { value: `per_min(count(${alias}))`, hint: 'count() / _minutes; avg in table' },
        { value: `per_sec(count(${alias}))`, hint: 'count() / _seconds; avg in table' },
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
