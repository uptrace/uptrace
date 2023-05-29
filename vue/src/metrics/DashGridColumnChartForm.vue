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
          :loading="timeseries.loading"
          class="mr-4"
          @click="timeseries.reload()"
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
        <v-row align="center">
          <v-col cols="auto">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">1</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="text-subtitle-1 text--primary">
              Select metrics that you want to plot or use in arithmetic expressions.
            </v-sheet>
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <MetricsPicker
              ref="metricsPicker"
              v-model="gridColumn.params.metrics"
              :table-grouping="tableGrouping"
              :uql="uql"
              :editable="editable"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-divider />
          </v-col>
        </v-row>

        <v-row align="center">
          <v-col cols="auto">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">2</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="text-subtitle-1 text--primary">
              Add some aggregations, filters, and groupings, for example,
              <code>per_min($metric_name{service.name='auth',region='eu'}) group by host.name</code
              >.
            </v-sheet>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricsQueryBuilder
              :date-range="dateRange"
              :metrics="activeMetrics"
              :uql="uql"
              :disabled="!activeMetrics.length"
              show-agg
              show-group-by
              show-dash-where
            />
          </v-col>
        </v-row>

        <v-row v-if="timeseries.status.hasData()">
          <v-col>
            <template v-if="gridColumn.params.columnMap">
              <v-chip
                v-for="(col, colName) in gridColumn.params.columnMap"
                :key="colName"
                outlined
                label
                class="ma-1"
              >
                <span>{{ colName }}</span>
                <UnitPicker v-model="col.unit" target-class="mr-n4" />
              </v-chip>
            </template>
          </v-col>
          <v-col cols="auto">
            <BtnSelectMenu v-model="gridColumn.params.chartKind" :items="chartKindItems" />
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <GridColumnChart
              :loading="timeseries.loading"
              :resolved="timeseries.status.isResolved()"
              :timeseries="styledTimeseries"
              :time="timeseries.time"
              :chart-kind="gridColumn.params.chartKind"
              :legend="gridColumn.params.legend"
            >
              <template v-if="editable" #expanded-item="{ headers, item, expandItem }">
                <tr class="v-data-table__expanded v-data-table__expanded__content">
                  <td :colspan="headers.length" class="py-2">
                    <TimeseriesStyleCard
                      :chart-kind="gridColumn.params.chartKind"
                      :timeseries-style="getTimeseriesStyle(item)"
                      @click:ok="expandItem(item, false)"
                      @click:reset="
                        expandItem(item, false)
                        resetTimeseriesStyle(item)
                      "
                    />
                  </td>
                </tr>
              </template>
            </GridColumnChart>
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-expansion-panels>
              <v-expansion-panel>
                <v-expansion-panel-header color="grey lighten-4"
                  >Chart Legend</v-expansion-panel-header
                >
                <v-expansion-panel-content class="pa-4">
                  <v-row align="center">
                    <v-col cols="3" class="text--secondary">Legend type</v-col>
                    <v-col cols="9">
                      <v-btn-toggle
                        v-model="gridColumn.params.legend.type"
                        color="deep-purple-accent-3"
                        group
                        dense
                      >
                        <v-btn :value="LegendType.None">None</v-btn>
                        <v-btn :value="LegendType.Table">Table</v-btn>
                        <v-btn :value="LegendType.List">List</v-btn>
                      </v-btn-toggle>
                    </v-col>
                  </v-row>

                  <v-row align="center">
                    <v-col cols="3" class="text--secondary">Legend placement</v-col>
                    <v-col cols="9">
                      <v-btn-toggle
                        v-model="gridColumn.params.legend.placement"
                        color="deep-purple-accent-3"
                        group
                        dense
                      >
                        <v-btn :value="LegendPlacement.Right">Right</v-btn>
                        <v-btn :value="LegendPlacement.Bottom">Bottom</v-btn>
                      </v-btn-toggle>
                    </v-col>
                  </v-row>

                  <v-row align="center">
                    <v-col cols="3" class="text--secondary">Legend values</v-col>
                    <v-col cols="9">
                      <v-select
                        v-model="gridColumn.params.legend.values"
                        multiple
                        :items="legendValueItems"
                        filled
                        dense
                        hide-details="auto"
                      />
                    </v-col>
                  </v-row>
                </v-expansion-panel-content>
              </v-expansion-panel>
            </v-expansion-panels>
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
                        placeholder="where host = ${host.name}"
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
import { defineComponent, shallowRef, set, del, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import { useGridColumnManager } from '@/metrics/use-dashboards'
import { useActiveMetrics } from '@/metrics/use-metrics'
import { useTimeseries, useStyledTimeseries } from '@/metrics/use-query'

// Components
import BtnSelectMenu from '@/components/BtnSelectMenu.vue'
import UnitPicker from '@/components/UnitPicker.vue'
import GridColumnChart from '@/metrics/GridColumnChart.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import TimeseriesStyleCard from '@/metrics/TimeseriesStyleCard.vue'

// Utilities
import { EventBus } from '@/models/eventbus'
import { requiredRule } from '@/util/validation'
import {
  defaultTimeseriesStyle,
  updateColumnMap,
  ChartGridColumn,
  ChartKind,
  Timeseries,
  StyledTimeseries,
  TimeseriesStyle,
  LegendType,
  LegendPlacement,
  LegendValue,
} from '@/metrics/types'

export default defineComponent({
  name: 'DashGridColumnChartForm',
  components: {
    BtnSelectMenu,
    GridColumnChart,
    UnitPicker,
    MetricsPicker,
    MetricsQueryBuilder,
    TimeseriesStyleCard,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    tableGrouping: {
      type: Array as PropType<string[]>,
      required: true,
    },
    gridColumn: {
      type: Object as PropType<ChartGridColumn>,
      required: true,
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const eventBus = new EventBus()
    const uql = useUql()

    const metricsPicker = shallowRef()
    const form = shallowRef()
    const isValid = shallowRef(false)
    const rules = { name: [requiredRule] }
    const gridColumnMan = useGridColumnManager()

    const chartKindItems = computed(() => {
      return [
        { text: 'Line chart', value: ChartKind.Line },
        { text: 'Area chart', value: ChartKind.Area },
        { text: 'Stacked area', value: ChartKind.StackedArea },
        { text: 'Bar chart', value: ChartKind.Bar },
        { text: 'Stacked bar', value: ChartKind.StackedBar },
      ]
    })
    const legendValueItems = computed(() => {
      return [
        { text: 'Avg', value: LegendValue.Avg },
        { text: 'Last', value: LegendValue.Last },
        { text: 'Min', value: LegendValue.Min },
        { text: 'Max', value: LegendValue.Max },
      ]
    })

    const activeMetrics = useActiveMetrics(computed(() => props.gridColumn.params.metrics))

    const timeseries = useTimeseries(() => {
      if (!props.gridColumn.params.metrics.length || !props.gridColumn.params.query) {
        return undefined
      }

      return {
        ...props.dateRange.axiosParams(),
        metric: props.gridColumn.params.metrics.map((m) => m.name),
        alias: props.gridColumn.params.metrics.map((m) => m.alias),
        query: props.gridColumn.params.query,
      }
    })

    const styledTimeseries = useStyledTimeseries(
      computed(() => timeseries.items),
      computed(() => props.gridColumn.params.columnMap),
      computed(() => props.gridColumn.params.timeseriesMap),
    )
    const currentTimeseries = shallowRef<StyledTimeseries[]>()
    const activeTimeseries = computed(() => {
      if (currentTimeseries.value) {
        return currentTimeseries.value
      }
      return styledTimeseries.value
    })

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
      () => timeseries.query,
      (query) => {
        if (query) {
          uql.setQueryInfo(query)
        }
      },
      { immediate: true },
    )

    watch(
      () => timeseries.columns,
      (columns) => {
        updateColumnMap(props.gridColumn.params.columnMap, columns)
      },
    )

    function submit() {
      const r1 = metricsPicker.value.validate()
      const r2 = form.value.validate()
      if (!r1 || !r2) {
        return
      }

      gridColumnMan.save(props.gridColumn).then((gridColumn) => {
        ctx.emit('click:save', gridColumn)
      })
    }

    function getTimeseriesStyle(ts: StyledTimeseries): TimeseriesStyle {
      if (!(ts.name in props.gridColumn.params.timeseriesMap)) {
        set(props.gridColumn.params.timeseriesMap, ts.name, {
          ...defaultTimeseriesStyle(),
          color: ts.color,
        })
      }
      return props.gridColumn.params.timeseriesMap[ts.name]
    }
    function resetTimeseriesStyle(ts: Timeseries) {
      del(props.gridColumn.params.timeseriesMap, ts.name)
    }

    return {
      LegendType,
      LegendPlacement,

      eventBus,
      uql,

      metricsPicker,
      form,
      isValid,
      rules,
      gridColumnMan,
      chartKindItems,
      legendValueItems,
      submit,

      activeMetrics,
      timeseries,
      styledTimeseries,
      currentTimeseries,
      activeTimeseries,

      getTimeseriesStyle,
      resetTimeseriesStyle,
    }
  },
})
</script>

<style lang="scss" scoped>
.v-expansion-panel-content ::v-deep .v-expansion-panel-content__wrap {
  padding: 0 !important;
}
</style>
