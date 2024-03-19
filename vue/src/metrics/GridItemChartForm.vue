<template>
  <GridItemFormPanes :grid-item="gridItem" v-on="$listeners">
    <template #picker>
      <MetricsPicker v-model="gridItem.params.metrics" :required-attrs="tableGrouping" :uql="uql" />
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
                show-agg
                show-group-by
                show-dash-where
              />
            </v-col>
          </v-row>

          <v-row>
            <v-col>
              <v-row v-if="Object.keys(gridItem.params.columnMap).length" class="mb-n6">
                <v-col>
                  <template v-if="gridItem.params.columnMap">
                    <v-chip
                      v-for="(col, colName) in gridItem.params.columnMap"
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
              </v-row>

              <v-row>
                <v-col>
                  <LegendaryChart
                    :loading="timeseries.loading"
                    :resolved="timeseries.status.isResolved()"
                    :timeseries="styledTimeseries"
                    :time="timeseries.time"
                    :chart-kind="gridItem.params.chartKind"
                    :legend="gridItem.params.legend"
                    :height="300"
                  ></LegendaryChart>
                </v-col>
              </v-row>
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
            v-model="gridItem.params.chartKind"
            :items="chartKindItems"
            label="Chart type"
            filled
            dense
          ></v-select>

          <v-checkbox
            v-model="gridItem.params.connectNulls"
            label="Connect the line across null points"
          />
        </SinglePanel>

        <SinglePanel title="Legend" expanded>
          <PanelSection title="Legend type" class="mb-4">
            <v-btn-toggle
              v-model="gridItem.params.legend.type"
              color="deep-purple-accent-3"
              group
              dense
              mandatory
            >
              <v-btn :value="LegendType.None">None</v-btn>
              <v-btn :value="LegendType.Table">Table</v-btn>
              <v-btn :value="LegendType.List">List</v-btn>
            </v-btn-toggle>
          </PanelSection>

          <PanelSection title="Legend placement">
            <v-btn-toggle
              v-model="gridItem.params.legend.placement"
              color="deep-purple-accent-3"
              group
              dense
              mandatory
            >
              <v-btn :value="LegendPlacement.Right">Right</v-btn>
              <v-btn :value="LegendPlacement.Bottom">Bottom</v-btn>
            </v-btn-toggle>
          </PanelSection>

          <PanelSection title="Legend values">
            <v-select
              v-model="gridItem.params.legend.values"
              multiple
              :items="legendValueItems"
              filled
              dense
              hide-details="auto"
              style="width: 200px"
            />
          </PanelSection>
        </SinglePanel>

        <SinglePanel
          v-for="ts in styledTimeseries"
          :key="ts.id"
          :title="`${ts.name} colum`"
          expanded
        >
          <TimeseriesStyleForm
            :chart-kind="gridItem.params.chartKind"
            :timeseries-style="getTimeseriesStyle(ts)"
          />
        </SinglePanel>
      </v-container>
    </template>
  </GridItemFormPanes>
</template>

<script lang="ts">
import { defineComponent, shallowRef, set, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql, joinQuery, injectQueryStore } from '@/use/uql'
import { useActiveMetrics } from '@/metrics/use-metrics'
import { useTimeseries, useStyledTimeseries } from '@/metrics/use-query'

// Components
import GridItemFormPanes from '@/metrics/GridItemFormPanes.vue'
import SinglePanel from '@/components/SinglePanel.vue'
import PanelSection from '@/components/PanelSection.vue'
import UnitPicker from '@/components/UnitPicker.vue'
import LegendaryChart from '@/metrics/LegendaryChart.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import TimeseriesStyleForm from '@/metrics/TimeseriesStyleForm.vue'

// Misc
import { requiredRule } from '@/util/validation'
import {
  defaultTimeseriesStyle,
  updateColumnMap,
  assignColors,
  emptyMetricColumn,
  ChartGridItem,
  ChartKind,
  StyledTimeseries,
  TimeseriesStyle,
  LegendType,
  LegendPlacement,
  LegendValue,
} from '@/metrics/types'

export default defineComponent({
  name: 'GridItemChartForm',
  components: {
    GridItemFormPanes,
    SinglePanel,
    PanelSection,
    LegendaryChart,
    UnitPicker,
    MetricsPicker,
    MetricsQueryBuilder,
    TimeseriesStyleForm,
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
    gridItem: {
      type: Object as PropType<ChartGridItem>,
      required: true,
    },
  },

  setup(props, ctx) {
    const uql = useUql()
    const rules = { title: [requiredRule] }

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

    const activeMetrics = useActiveMetrics(computed(() => props.gridItem.params.metrics))

    const { where } = injectQueryStore()
    const timeseries = useTimeseries(() => {
      if (!props.gridItem.params.metrics.length || !props.gridItem.params.query) {
        return undefined
      }

      return {
        ...props.dateRange.axiosParams(),
        metric: props.gridItem.params.metrics.map((m) => m.name),
        alias: props.gridItem.params.metrics.map((m) => m.alias),
        query: joinQuery([props.gridItem.params.query, where.value]),
      }
    })

    const styledTimeseries = useStyledTimeseries(
      computed(() => timeseries.items),
      computed(() => props.gridItem.params.columnMap),
      computed(() => props.gridItem.params.timeseriesMap),
    )
    const currentTimeseries = shallowRef<StyledTimeseries[]>()
    const activeTimeseries = computed(() => {
      if (currentTimeseries.value) {
        return currentTimeseries.value
      }
      return styledTimeseries.value
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
        updateColumnMap(props.gridItem.params.columnMap, columns, emptyMetricColumn)
        assignColors(props.gridItem.params.columnMap, columns)
      },
    )

    function getTimeseriesStyle(ts: StyledTimeseries): TimeseriesStyle {
      if (!(ts.name in props.gridItem.params.timeseriesMap)) {
        set(props.gridItem.params.timeseriesMap, ts.name, {
          ...defaultTimeseriesStyle(),
          color: ts.color,
        })
      }
      return props.gridItem.params.timeseriesMap[ts.name]
    }

    return {
      LegendType,
      LegendPlacement,

      uql,

      rules,
      chartKindItems,
      legendValueItems,

      activeMetrics,
      timeseries,
      styledTimeseries,
      currentTimeseries,
      activeTimeseries,

      getTimeseriesStyle,
    }
  },
})
</script>

<style lang="scss" scoped></style>
