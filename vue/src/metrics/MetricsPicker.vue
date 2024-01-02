<template>
  <v-card flat>
    <v-row align="center" class="mb-n6">
      <v-col cols="auto" class="text-subtitle-1 text--secondary">
        Active metrics
        <v-btn
          v-if="!attrFilterEnabled"
          icon
          title="Show metrics filter"
          class="ml-2"
          @click="attrFilterEnabled = true"
        >
          <v-icon>mdi-filter</v-icon>
        </v-btn>
      </v-col>
      <v-col v-if="attrFilterEnabled" cols="auto">
        <v-autocomplete
          v-model="activeAttrKeys"
          multiple
          :loading="attrKeysDs.loading"
          :items="attrKeysDs.filteredItems"
          :error-messages="attrKeysDs.errorMessages"
          :search-input.sync="attrKeysDs.searchInput"
          label="Show all metrics"
          placeholder="Show metrics with given attributes..."
          solo
          flat
          dense
          background-color="grey lighten-4"
          no-filter
          auto-select-first
          clearable
          hide-details="auto"
          style="min-width: 320px"
        >
          <template #item="{ item }">
            <v-list-item-action class="my-0 mr-4">
              <v-simple-checkbox :value="activeAttrKeys.includes(item.value)"></v-simple-checkbox>
            </v-list-item-action>
            <v-list-item-content>
              <v-list-item-title>
                {{ item.text }}
              </v-list-item-title>
            </v-list-item-content>
            <v-list-item-action>
              <v-chip small>{{ item.count }}</v-chip>
            </v-list-item-action>
          </template>
        </v-autocomplete>
      </v-col>
    </v-row>

    <v-row v-if="!activeMetrics.length" align="center" class="text-body-2">
      <v-col cols="auto"><v-icon color="orange">mdi-lightbulb</v-icon></v-col>
      <v-col>
        Select a metric, specify a short alias, and click the "Apply" button to add default
        aggregations.
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <MetricPicker
          v-for="(metric, i) in value"
          :key="`${metric.alias}-${i}`"
          :loading="metrics.loading"
          :metrics="metrics.items"
          :value="metric"
          :active-metrics="value"
          :query="uql.query"
          @click:apply="updateMetric(metric, $event)"
          @click:remove="removeMetric(i, metric)"
        />
        <MetricPicker
          v-if="activeMetrics.length < 10"
          :loading="metrics.loading"
          :metrics="metrics.items"
          :active-metrics="value"
          :query="uql.query"
          @click:apply="addMetric($event)"
        />
      </v-col>
    </v-row>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useDataSource } from '@/use/datasource'
import { injectForceReload } from '@/use/force-reload'
import { createQueryEditor, UseUql } from '@/use/uql'
import { useMetrics, useActiveMetrics, defaultMetricQuery } from '@/metrics/use-metrics'
import { MetricAlias } from '@/metrics/types'

// Components
import MetricPicker from '@/metrics/MetricPicker.vue'

// Misc
import { escapeRe } from '@/util/string'

export default defineComponent({
  name: 'MetricsPicker',
  components: { MetricPicker },

  props: {
    value: {
      type: Array as PropType<MetricAlias[]>,
      required: true,
    },
    requiredAttrs: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    autoGrouping: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const route = useRoute()
    const forceReload = injectForceReload()

    const attrFilterEnabled = shallowRef(false)
    const activeAttrKeys = shallowRef<string[]>([])
    watch(
      () => props.requiredAttrs,
      (grouping) => {
        activeAttrKeys.value = grouping
      },
      { immediate: true },
    )

    const attrKeysDs = useDataSource(() => {
      const { projectId } = route.value.params
      return {
        url: `/internal/v1/metrics/${projectId}/attr-keys`,
        params: {
          ...forceReload.params,
        },
      }
    })
    const metrics = useMetrics(() => {
      return {
        attr_key: activeAttrKeys.value,
      }
    })
    const activeMetrics = useActiveMetrics(computed(() => props.value))

    function addMetric(newMetric: MetricAlias) {
      updateQuery({ name: '', alias: '' }, newMetric)

      const activeMetrics = props.value.slice()
      activeMetrics.push(newMetric)
      ctx.emit('input', activeMetrics)
    }

    function updateMetric(oldMetric: MetricAlias, newMetric: MetricAlias) {
      updateQuery(oldMetric, newMetric)

      oldMetric.name = newMetric.name
      oldMetric.alias = newMetric.alias
    }

    function updateQuery(oldMetric: MetricAlias, newMetric: MetricAlias) {
      if (oldMetric.alias) {
        const re = createRegexp(oldMetric.alias)
        if (re.test(props.uql.query)) {
          props.uql.query = props.uql.query.replaceAll(
            createRegexp(oldMetric.alias, 'g'),
            '$' + newMetric.alias,
          )
          return
        }
      }

      const metric = metrics.items.find((m) => m.name === newMetric.name)
      if (!metric) {
        return
      }

      const editor = createQueryEditor(props.uql.query)

      if (props.autoGrouping && !props.uql.query) {
        for (let attrKey of metric.attrKeys) {
          editor.groupBy(attrKey)
        }
      }

      const column = defaultMetricQuery(metric.instrument, newMetric.alias)
      editor.add(column)

      props.uql.query = editor.toString()
    }

    function removeMetric(index: number, metric: MetricAlias) {
      const re = createRegexp(metric.alias)
      props.uql.parts = props.uql.parts.filter((part) => {
        return !re.test(part.query)
      })

      const activeMetrics = props.value.slice()
      activeMetrics.splice(index, 1)
      ctx.emit('input', activeMetrics)
    }

    return {
      attrFilterEnabled,
      activeAttrKeys,
      attrKeysDs,
      metrics,

      activeMetrics,
      addMetric,
      updateMetric,
      removeMetric,
    }
  },
})

function createRegexp(alias: string, flags = '') {
  return new RegExp(escapeRe('$' + alias) + '\\b', flags)
}
</script>

<style lang="scss" scoped></style>
