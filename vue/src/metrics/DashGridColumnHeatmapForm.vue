<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-card>
      <v-toolbar color="light-blue lighten-5" flat>
        <v-toolbar-title>
          {{ gridColumn.id ? 'Edit chart column' : 'New chart column' }}
        </v-toolbar-title>
        <v-btn icon href="https://uptrace.dev/get/querying-metrics.html" target="_blank"
          ><v-icon>mdi-help-circle-outline</v-icon></v-btn
        >

        <v-spacer />

        <v-btn
          small
          outlined
          :loading="heatmapQuery.loading"
          class="mr-4"
          @click="heatmapQuery.reload()"
        >
          <v-icon small left>mdi-refresh</v-icon>
          <span>Reload</span>
        </v-btn>

        <v-toolbar-items>
          <v-btn icon @click="$emit('click:cancel')">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>
      </v-toolbar>

      <v-container fluid class="pa-6">
        <v-row>
          <v-col cols="3" class="mt-4 text--secondary">Histogram metric</v-col>
          <v-col cols="6">
            <v-autocomplete
              v-model="gridColumn.params.metric"
              :items="heatmapMetrics"
              item-text="name"
              item-value="name"
              placeholder="Select a histogram metric..."
              filled
              dense
              :rules="rules.metric"
              hide-details="auto"
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

        <v-row>
          <v-col cols="3" class="mt-4 text--secondary">Metric unit</v-col>
          <v-col cols="3">
            <v-select
              v-model="gridColumn.params.unit"
              :items="unitItems"
              filled
              dense
              hide-details="auto"
            ></v-select>
          </v-col>
        </v-row>

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
              :unit="gridColumn.params.unit"
              :x-axis="heatmapQuery.xAxis"
              :y-axis="heatmapQuery.yAxis"
              :data="heatmapQuery.data"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-divider />
          </v-col>
        </v-row>

        <v-row>
          <v-col cols="3" class="mt-4 text--secondary">Chart name</v-col>
          <v-col cols="9">
            <v-text-field
              v-model="gridColumn.name"
              hint="Short name that describes the chart"
              persistent-hint
              filled
              :rules="rules.name"
              hide-details="auto"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col cols="3" class="mt-4 text--secondary">Optional description</v-col>
          <v-col cols="9">
            <v-text-field
              v-model="gridColumn.description"
              hint="Optional description or comment"
              persistent-hint
              filled
              hide-details="auto"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-expansion-panels>
              <v-expansion-panel>
                <v-expansion-panel-header color="grey lighten-4">
                  Advanced
                </v-expansion-panel-header>
                <v-expansion-panel-content class="pa-4">
                  <v-row>
                    <v-col cols="3" class="mt-4 text--secondary">Grid query template</v-col>
                    <v-col cols="9">
                      <v-text-field
                        v-model="gridColumn.gridQueryTemplate"
                        hint="Template"
                        persistent-hint
                        filled
                        hide-details="auto"
                      />
                    </v-col>
                  </v-row>
                </v-expansion-panel-content>
              </v-expansion-panel>
            </v-expansion-panels>
          </v-col>
        </v-row>

        <v-row v-if="editable" class="mt-8">
          <v-col>
            <v-divider />
          </v-col>
        </v-row>

        <v-row v-if="editable">
          <v-spacer />
          <v-col cols="auto">
            <v-btn text class="mr-2" @click="$emit('click:cancel')">Cancel</v-btn>
            <v-btn
              type="submit"
              color="primary"
              :disabled="!isValid"
              :loading="gridColumnMan.pending"
              >{{ gridColumn.id ? 'Save' : 'Create' }}</v-btn
            >
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import { useActiveMetrics } from '@/metrics/use-metrics'
import { useHeatmapQuery } from '@/metrics/use-query'
import { useGridColumnManager } from '@/metrics/use-dashboards'

// Components
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import HeatmapChart from '@/components/HeatmapChart.vue'

// Utilities
import { UNITS } from '@/util/fmt'
import { requiredRule } from '@/util/validation'
import { HeatmapGridColumn, Metric, Instrument } from '@/metrics/types'

export default defineComponent({
  name: 'DashGridColumnHeatmapForm',
  components: { MetricsQueryBuilder, HeatmapChart },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    metrics: {
      type: Array as PropType<Metric[]>,
      required: true,
    },
    gridColumn: {
      type: Object as PropType<HeatmapGridColumn>,
      required: true,
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const form = shallowRef()
    const isValid = shallowRef(false)
    const rules = { metric: [requiredRule], name: [requiredRule] }
    const gridColumnMan = useGridColumnManager()

    const unitItems = computed(() => {
      return UNITS.map((unit) => {
        return { value: unit, text: unit || 'none' }
      })
    })

    watch(
      () => props.gridColumn.params.metric,
      (metricName) => {
        const metric = heatmapMetrics.value.find((metric) => metric.name === metricName)
        if (metric) {
          props.gridColumn.params.unit = metric.unit
        }
      },
    )

    function submit() {
      if (!form.value.validate()) {
        return
      }

      gridColumnMan.save(props.gridColumn).then((gridColumn) => {
        ctx.emit('click:save', gridColumn)
      })
    }

    const heatmapMetrics = computed(() => {
      return props.metrics.filter((metric) => {
        return metric.instrument === Instrument.Histogram
      })
    })

    const heatmapQuery = useHeatmapQuery(() => {
      if (!props.gridColumn.params.metric) {
        return undefined
      }

      return {
        ...props.dateRange.axiosParams(),
        metric: props.gridColumn.params.metric,
        alias: props.gridColumn.params.metric,
        query: props.gridColumn.params.query,
      }
    })

    const uql = useUql()
    const activeMetrics = useActiveMetrics(
      computed(() => [
        {
          name: props.gridColumn.params.metric,
          alias: props.gridColumn.params.metric,
        },
      ]),
    )

    watch(
      () => props.gridColumn.params.query,
      (query) => {
        uql.query = query
      },
      { immediate: true },
    )

    watch(
      () => uql.query,
      (query) => {
        props.gridColumn.params.query = query
      },
    )

    watch(
      () => heatmapQuery.query,
      (query) => {
        if (query) {
          uql.setQueryInfo(query)
        }
      },
      { immediate: true },
    )

    return {
      Instrument,

      form,
      isValid,
      rules,
      gridColumnMan,
      unitItems,
      submit,

      heatmapMetrics,
      heatmapQuery,

      uql,
      activeMetrics,
    }
  },
})
</script>

<style lang="scss" scoped></style>
