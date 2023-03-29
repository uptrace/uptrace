<template>
  <div>
    <v-row v-for="(metric, index) in activeMetrics" :key="metric.alias" dense align="center">
      <v-col v-if="editable" cols="auto">
        <v-btn icon title="Remove metric" @click="removeMetric(index, metric)">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-col>
      <v-col cols="auto">
        <span :title="metric.description" class="font-weight-bold">${{ metric.alias }}</span>
        <span class="mx-2">({{ metric.name }})</span>
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

    <v-row v-if="editable">
      <v-col>
        <v-card outlined rounded="lg" class="pa-4">
          <v-row align="center">
            <v-col cols="auto" class="text-body-2 text--secondary"
              >Only show metrics with attributes</v-col
            >
            <v-col cols="auto">
              <v-autocomplete
                v-model="activeAttrKeys"
                multiple
                :loading="attrKeysDs.loading"
                :items="attrKeysDs.filteredItems"
                :error-messages="attrKeysDs.errorMessages"
                :search-input.sync="attrKeysDs.searchInput"
                placeholder="Show all metrics"
                solo
                flat
                dense
                background-color="grey lighten-4"
                no-filter
                auto-select-first
                clearable
                hide-details="auto"
              >
                <template #item="{ item }">
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

          <v-row>
            <v-col>
              <MetricPicker
                v-if="value.length < 6"
                ref="metricPicker"
                :loading="metrics.loading"
                :table-grouping="tableGrouping"
                :metrics="metrics.items"
                :active-metrics="value"
                :uql="uql"
                :required="value.length === 0"
                @click:add="addMetric($event)"
              />
            </v-col>
          </v-row>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useDataSource } from '@/use/datasource'
import { useForceReload } from '@/use/force-reload'
import { UseUql } from '@/use/uql'
import { useMetrics, useActiveMetrics, defaultMetricQuery } from '@/metrics/use-metrics'
import { hasMetricAlias } from '@/metrics/use-query'
import { MetricAlias } from '@/metrics/types'

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
    tableGrouping: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const route = useRoute()
    const { forceReloadParams } = useForceReload()

    const metricPicker = shallowRef()
    function validate(): boolean {
      if (!metricPicker.value) {
        return true
      }
      return metricPicker.value.validate()
    }

    const activeAttrKeys = shallowRef<string[]>([])
    watch(
      () => props.tableGrouping,
      (grouping) => {
        activeAttrKeys.value = grouping
      },
      { immediate: true },
    )

    const attrKeysDs = useDataSource(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/metrics/${projectId}/attr-keys`,
        params: {
          ...forceReloadParams.value,
        },
      }
    })
    const metrics = useMetrics(() => {
      return {
        attr_key: activeAttrKeys.value,
      }
    })
    const activeMetrics = useActiveMetrics(computed(() => props.value))

    function addMetric(metricAlias: MetricAlias) {
      const metric = metrics.items.find((m) => m.name === metricAlias.name)
      if (!metric) {
        return
      }

      const activeMetrics = props.value.slice()
      activeMetrics.push(metricAlias)
      ctx.emit('input', activeMetrics)

      const column = defaultMetricQuery(metric.instrument, metricAlias.alias)
      props.uql.query = props.uql.query + ' | ' + column
    }

    function removeMetric(index: number, metric: MetricAlias) {
      const activeMetrics = props.value.slice()
      activeMetrics.splice(index, 1)
      ctx.emit('input', activeMetrics)

      props.uql.parts = props.uql.parts.filter((part) => {
        return !hasMetricAlias(part.query, metric.alias)
      })
    }

    return {
      activeAttrKeys,
      attrKeysDs,
      metrics,

      activeMetrics,
      addMetric,
      removeMetric,

      metricPicker,
      validate,
    }
  },
})
</script>

<style lang="scss" scoped></style>
