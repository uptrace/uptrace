<template>
  <v-form v-model="form.isValid">
    <v-card outlined>
      <v-toolbar color="light-blue lighten-5" flat>
        <v-toolbar-title>Dashboard entry</v-toolbar-title>
        <v-btn icon href="https://uptrace.dev/docs/querying-metrics.html" target="_blank"
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

      <v-container class="py-6 indent-rows">
        <v-row align="center" dense>
          <v-col cols="auto" class="pr-4">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">1</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="text-subtitle-1 text--primary">
              Select up to 5 metrics that you want to plot or use in arithmetic expressions.
            </v-sheet>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricsPicker v-model="dashEntry.metrics" :metrics="metrics.items" :uql="uql" />
          </v-col>
        </v-row>

        <v-divider class="my-8" />

        <v-row align="center" dense>
          <v-col cols="auto" class="pr-4">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">2</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="text-subtitle-1 text--primary">
              Select some aggregations, filters, and groupings, for example,
              <code>per_min($metric_name{key1=value1,key2=value2}) group by host.name</code>.
            </v-sheet>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricQueryBuilder
              :date-range="dateRange"
              :metrics="activeMetrics"
              :uql="uql"
              :disabled="!activeMetrics.length"
              show-metric-group-by
            />
          </v-col>
        </v-row>

        <v-divider class="my-8" />

        <v-row align="center">
          <v-col cols="5">
            <v-text-field
              v-model="dashEntry.name"
              label="Chart title"
              solo
              flat
              background-color="grey lighten-4"
              :rules="form.rules.name"
              required
              hide-details
            />
          </v-col>
          <v-col>
            <v-text-field
              v-model="dashEntry.description"
              label="Optional description or comment"
              placeholder="Optional description or comment"
              solo
              flat
              background-color="grey lighten-4"
              hide-details
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-chip v-for="(col, colName) in columnMap" :key="colName" outlined label class="ma-1">
              <span>{{ colName }}</span>
              <UnitPicker v-model="col.unit" target-class="mr-n4" />
            </v-chip>
          </v-col>
          <v-col cols="auto">
            <BtnSelectMenu v-model="dashEntry.chartType" :items="form.chartTypeItems" />
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricChart
              :loading="timeseries.loading"
              :resolved="timeseries.status.isResolved()"
              :chart-type="dashEntry.chartType"
              :column-map="columnMap"
              :timeseries="timeseries.items"
              show-legend
            />
          </v-col>
        </v-row>

        <v-divider class="my-8" />

        <v-row v-if="!dashboard.isTemplate">
          <v-spacer />
          <v-col cols="auto">
            <v-btn text class="mr-2" @click="$emit('click:cancel')">Cancel</v-btn>
            <v-btn
              color="primary"
              :disabled="!form.isValid"
              :loading="dashEntryMan.pending"
              @click="onClickSave"
              >Save</v-btn
            >
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-form>
</template>

<script lang="ts">
import {
  defineComponent,
  shallowRef,
  reactive,
  computed,
  watch,
  onMounted,
  proxyRefs,
  PropType,
} from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import { useDashEntryManager, UseDashboard, DashEntry, ChartType } from '@/metrics/use-dashboards'
import { useActiveMetrics, UseMetrics, MetricColumn } from '@/metrics/use-metrics'
import { UseTimeseries } from '@/metrics/use-query'

// Components
import BtnSelectMenu from '@/components/BtnSelectMenu.vue'
import UnitPicker from '@/components/UnitPicker.vue'
import MetricChart from '@/metrics/MetricChart.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricQueryBuilder from '@/metrics/query/MetricQueryBuilder.vue'

// Utilities
import { requiredRule } from '@/util/validation'

export default defineComponent({
  name: 'DashEntryForm',
  components: {
    BtnSelectMenu,
    MetricChart,
    UnitPicker,
    MetricsPicker,
    MetricQueryBuilder,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    metrics: {
      type: Object as PropType<UseMetrics>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<UseDashboard>,
      required: true,
    },
    dashEntry: {
      type: Object as PropType<DashEntry>,
      required: true,
    },
    timeseries: {
      type: Object as PropType<UseTimeseries>,
      required: true,
    },
  },

  setup(props, ctx) {
    const uql = useUql()
    const dashEntryMan = useDashEntryManager()

    const activeMetrics = useActiveMetrics(
      computed(() => props.metrics.items),
      computed(() => props.dashEntry.metrics),
    )

    const columnMap = computed((): Record<string, MetricColumn> => {
      const columnMap = {}

      for (let ts of props.timeseries.items) {
        columnMap[ts.metric] = {
          unit: ts.unit,
        }
      }
      for (let colName in props.dashEntry.columnMap) {
        if (colName in columnMap) {
          columnMap[colName] = props.dashEntry.columnMap[colName]
        }
      }

      return reactive(columnMap)
    })

    onMounted(() => {
      props.metrics.reload()
    })

    watch(
      () => props.dashEntry.query,
      (query) => {
        uql.query = query
      },
      { immediate: true },
    )

    watch(
      () => uql.query,
      (query) => {
        props.dashEntry.query = query
      },
    )

    watch(
      () => props.timeseries.queryParts,
      (queryParts) => {
        if (queryParts) {
          uql.syncParts(queryParts)
        }
      },
      { immediate: true },
    )

    function onClickSave() {
      dashEntryMan.save({ ...props.dashEntry, columnMap: columnMap.value }).then((dashEntry) => {
        ctx.emit('click:save', dashEntry)
      })
    }

    return {
      uql,
      dashEntryMan,

      activeMetrics,
      columnMap,
      form: useForm(),

      onClickSave,
    }
  },
})

function useForm() {
  const isValid = shallowRef(false)
  const rules = { name: [requiredRule] }

  const chartTypeItems = computed(() => {
    return [
      { text: 'Line chart', value: ChartType.Line },
      { text: 'Area chart', value: ChartType.Area },
      { text: 'Stacked area', value: ChartType.StackedArea },
      { text: 'Bar chart', value: ChartType.Bar },
      { text: 'Stacked bar', value: ChartType.StackedBar },
    ]
  })

  return proxyRefs({
    isValid,
    rules,
    chartTypeItems,
  })
}
</script>

<style lang="scss" scoped>
.indent-rows ::v-deep .row {
  padding-left: 12px !important;
  padding-right: 12px !important;
}
</style>
