<template>
  <GridItemFormPanes :grid-item="gridItem" v-on="$listeners">
    <template #picker>
      <v-container fluid class="py-16">
        <v-row justify="center" align="center">
          <v-col cols="auto">
            <v-autocomplete
              v-model="gridItem.params.metric"
              :items="metrics.items"
              item-text="name"
              item-value="name"
              label="Histogram metric"
              filled
              dense
              :rules="rules.metric"
              hide-details="auto"
              style="min-width: 600px"
            >
              <template #item="{ item }">
                <v-list-item-content>
                  <v-list-item-title>
                    <span>{{ item.name }}</span>
                    <v-chip label small color="grey lighten-4" title="Instrument" class="ml-2">{{
                      item.instrument
                    }}</v-chip>
                    <v-chip
                      v-if="item.unit"
                      label
                      small
                      color="grey lighten-4"
                      title="Unit"
                      class="ml-2"
                      >{{ item.unit }}</v-chip
                    >
                  </v-list-item-title>
                  <v-list-item-subtitle>
                    {{ item.description }}
                  </v-list-item-subtitle>
                </v-list-item-content>
              </template>
            </v-autocomplete>
          </v-col>
        </v-row>
      </v-container>
    </template>
    <template #preview>
      <v-container fluid>
        <v-row v-if="!activeMetrics.length">
          <v-col>
            <v-skeleton-loader type="image" boilerplate></v-skeleton-loader>
          </v-col>
        </v-row>

        <template v-else>
          <v-row>
            <v-col>
              <MetricsQueryBuilder
                :date-range="dateRange"
                :metrics="activeMetrics"
                :uql="uql"
                :disabled="!activeMetrics.length"
                show-dash-where
              />
            </v-col>
          </v-row>

          <v-row>
            <v-col>
              <HeatmapChart
                :loading="heatmapQuery.loading"
                :resolved="heatmapQuery.status.isResolved()"
                :unit="gridItem.params.unit"
                :x-axis="heatmapQuery.xAxis"
                :y-axis="heatmapQuery.yAxis"
                :data="heatmapQuery.data"
              />
            </v-col>
          </v-row>
        </template>
      </v-container>
    </template>
    <template #options>
      <v-container fluid>
        <SinglePanel title="Chart options" expanded>
          <v-text-field
            v-model="gridItem.title"
            label="Chart title"
            filled
            dense
            :rules="rules.title"
          />

          <v-text-field
            v-model="gridItem.description"
            label="Optional description or memo"
            filled
            dense
          />

          <v-select
            v-model="gridItem.params.unit"
            label="Metric unit"
            :items="unitItems"
            filled
            dense
            hide-details="auto"
          ></v-select>
        </SinglePanel>
      </v-container>
    </template>
  </GridItemFormPanes>
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql, joinQuery, injectQueryStore } from '@/use/uql'
import { useMetrics, useActiveMetrics } from '@/metrics/use-metrics'
import { useHeatmapQuery } from '@/metrics/use-query'

// Components
import GridItemFormPanes from '@/metrics/GridItemFormPanes.vue'
import SinglePanel from '@/components/SinglePanel.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import HeatmapChart from '@/components/HeatmapChart.vue'

// Misc
import { UNITS } from '@/util/fmt'
import { requiredRule } from '@/util/validation'
import { HeatmapGridItem, Instrument } from '@/metrics/types'

export default defineComponent({
  name: 'GridItemHeatmapForm',
  components: { GridItemFormPanes, SinglePanel, MetricsQueryBuilder, HeatmapChart },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    gridItem: {
      type: Object as PropType<HeatmapGridItem>,
      required: true,
    },
  },

  setup(props, ctx) {
    const rules = { metric: [requiredRule], title: [requiredRule] }

    const uql = useUql()
    const { where } = injectQueryStore()
    const activeMetrics = useActiveMetrics(
      computed(() => {
        if (!props.gridItem.params.metric) {
          return []
        }
        return [
          {
            name: props.gridItem.params.metric,
            alias: props.gridItem.params.metric,
            query: joinQuery([props.gridItem.params.query, where.value]),
          },
        ]
      }),
    )

    const unitItems = computed(() => {
      return UNITS.map((unit) => {
        return { value: unit, text: unit || 'none' }
      })
    })

    const metrics = useMetrics(() => {
      return { instrument: Instrument.Histogram }
    })

    watch(
      () => props.gridItem.params.metric,
      (metricName) => {
        const metric = metrics.items.find((metric) => metric.name === metricName)
        if (metric) {
          props.gridItem.params.unit = metric.unit
        }
      },
    )

    const heatmapQuery = useHeatmapQuery(() => {
      if (!props.gridItem.params.metric) {
        return undefined
      }

      return {
        ...props.dateRange.axiosParams(),
        metric: props.gridItem.params.metric,
        alias: props.gridItem.params.metric,
        query: props.gridItem.params.query,
      }
    })

    watch(
      () => props.gridItem.params.query,
      (query) => {
        uql.query = query
      },
      { immediate: true },
    )
    watch(
      () => uql.query,
      (query) => {
        props.gridItem.params.query = query
      },
    )

    watch(
      () => heatmapQuery.query,
      (query) => {
        if (query) {
          uql.setQueryInfo(query)
        }
      },
    )

    return {
      Instrument,

      rules,
      unitItems,

      metrics,
      heatmapQuery,

      uql,
      activeMetrics,
    }
  },
})
</script>

<style lang="scss" scoped></style>
