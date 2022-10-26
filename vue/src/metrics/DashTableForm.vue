<template>
  <v-form v-model="form.isValid">
    <v-card outlined>
      <v-toolbar color="light-blue lighten-5" flat>
        <v-toolbar-title>Dashboard table</v-toolbar-title>
        <v-btn icon href="https://uptrace.dev/get/querying-metrics.html" target="_blank"
          ><v-icon>mdi-help-circle-outline</v-icon></v-btn
        >
        <DashLockedIcon v-if="dashboard.isTemplate" />

        <v-spacer />

        <v-btn
          small
          outlined
          :loading="tableQuery.loading"
          class="mr-4"
          @click="tableQuery.reload()"
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
          <v-col cols="auto">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">1</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="px-4 text-subtitle-1 text--primary">
              Select up to 6 metrics you want to display for each row in the table. The selected
              metrics should have some common attributes that will be used to join metrics together.
            </v-sheet>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricsPicker
              v-model="dashboard.metrics"
              :metrics="metrics.items"
              :uql="uql"
              :disabled="dashboard.isGrid"
            />
          </v-col>
        </v-row>

        <v-divider class="my-8" />

        <v-row align="center" dense>
          <v-col cols="auto">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">2</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="px-4 text-subtitle-1 text--primary">
              Select some aggregations and group-by attributes to display as columns in the table.
              Each row in the table will lead to a separate grid-based dashboard.
            </v-sheet>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricQueryBuilder
              :date-range="dateRange"
              :metrics="activeMetrics"
              :uql="uql"
              show-dash-group-by
              :disabled="!activeMetrics.length"
            />
          </v-col>
        </v-row>

        <v-divider class="my-8" />

        <v-row>
          <v-col>
            <v-row dense>
              <v-col cols="auto">
                <v-chip
                  v-for="(col, colName) in columnMap"
                  :key="colName"
                  outlined
                  label
                  class="ma-1"
                >
                  <span>{{ colName }}</span>
                  <UnitPicker v-model="col.unit" target-class="mr-n4" />
                </v-chip>
              </v-col>
            </v-row>

            <v-row dense>
              <v-col>
                <MetricItemsTable
                  :loading="tableQuery.loading"
                  :items="tableQuery.items"
                  :columns="tableQuery.columns"
                  :order="tableQuery.order"
                  :axios-params="axiosParams"
                  :column-map="columnMap"
                />
              </v-col>
            </v-row>
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
              :loading="dashMan.pending"
              @click="saveDash"
              >Save</v-btn
            >
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-form>
</template>

<script lang="ts">
import { defineComponent, shallowRef, reactive, computed, watch, proxyRefs, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import { useActiveMetrics, UseMetrics } from '@/metrics/use-metrics'
import { useDashManager, UseDashboard } from '@/metrics/use-dashboards'
import { UseTableQuery } from '@/metrics/use-query'
import { MetricColumn } from '@/metrics/types'

// Components
import UnitPicker from '@/components/UnitPicker.vue'
import DashLockedIcon from '@/metrics/DashLockedIcon.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricQueryBuilder from '@/metrics/query/MetricQueryBuilder.vue'
import MetricItemsTable from '@/metrics/MetricItemsTable.vue'

export default defineComponent({
  name: 'DashTableForm',
  components: {
    UnitPicker,
    DashLockedIcon,
    MetricsPicker,
    MetricQueryBuilder,
    MetricItemsTable,
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
    tableQuery: {
      type: Object as PropType<UseTableQuery>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      default: undefined,
    },
  },

  setup(props, ctx) {
    const uql = useUql()
    const dashMan = useDashManager()

    const activeMetrics = useActiveMetrics(
      computed(() => props.metrics.items),
      computed(() => props.dashboard.metrics),
    )

    const columnMap = computed((): Record<string, MetricColumn> => {
      const columnMap: Record<string, MetricColumn> = {}

      for (let col of props.tableQuery.columns) {
        if (!col.isGroup) {
          columnMap[col.name] = {
            unit: col.unit,
          }
        }
      }
      for (let colName in props.dashboard.columnMap) {
        if (colName in columnMap) {
          columnMap[colName] = props.dashboard.columnMap[colName]
        }
      }

      return reactive(columnMap)
    })

    watch(
      () => props.dashboard.data?.query ?? '',
      (query) => {
        uql.query = query
      },
      { immediate: true },
    )

    watch(
      () => uql.query,
      (query) => {
        if (props.dashboard.data) {
          props.dashboard.data.query = query
        }
      },
    )

    watch(
      () => props.tableQuery.queryParts,
      (queryParts) => {
        if (queryParts) {
          uql.syncParts(queryParts)
        }
      },
      { immediate: true },
    )

    function saveDash() {
      if (!props.dashboard.data) {
        return
      }

      dashMan
        .update({
          baseQuery: props.dashboard.data.baseQuery,
          metrics: props.dashboard.data.metrics,
          query: props.dashboard.data.query,
          columnMap: columnMap.value,
        })
        .then((dash) => {
          ctx.emit('click:save', dash)
        })
    }

    return {
      uql,
      dashMan,
      form: useForm(),

      activeMetrics,
      columnMap,

      saveDash,
    }
  },
})

function useForm() {
  const isValid = shallowRef(false)

  return proxyRefs({
    isValid,
  })
}
</script>

<style lang="scss" scoped>
.indent-rows ::v-deep .row {
  padding-left: 12px !important;
  padding-right: 12px !important;
}
</style>
