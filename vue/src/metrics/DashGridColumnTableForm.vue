<template>
  <v-form ref="form" v-model="isValid" lazy-validation @submit.prevent="submit">
    <v-card outlined>
      <v-toolbar color="light-blue lighten-5" flat>
        <v-toolbar-title>
          {{ gridColumn.id ? 'Edit table column' : 'New table column' }}
        </v-toolbar-title>
        <v-btn icon href="https://uptrace.dev/get/querying-metrics.html" target="_blank"
          ><v-icon>mdi-help-circle-outline</v-icon></v-btn
        >

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

      <v-container class="pa-6">
        <v-row align="center" dense>
          <v-col cols="auto">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">1</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="px-4 text-subtitle-1 text--primary">
              Select metrics you want to display for each row in the table. The selected metrics
              should have some common attributes that will be used to join timeseries together.
            </v-sheet>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricsPicker
              ref="metricsPicker"
              v-model="gridColumn.params.metrics"
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

        <v-row align="center" dense>
          <v-col cols="auto">
            <v-avatar color="blue darken-1" size="40">
              <span class="white--text text-h5">2</span>
            </v-avatar>
          </v-col>
          <v-col>
            <v-sheet max-width="800" class="px-4 text-subtitle-1 text--primary">
              Add some aggregations and group-by attributes to display as columns in the table. Each
              row in the table will lead to a separate grid-based dashboard.
            </v-sheet>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <MetricsQueryBuilder
              :date-range="dateRange"
              :metrics="activeMetrics"
              :uql="uql"
              show-agg
              show-group-by
              show-metrics-where
              :disabled="!activeMetrics.length"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-divider />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-row dense>
              <v-col cols="auto">
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
              </v-col>
            </v-row>

            <v-row dense>
              <v-col>
                <TimeseriesTable
                  :loading="tableQuery.loading"
                  :items="tableQuery.items"
                  :columns="tableQuery.columns"
                  :order="tableQuery.order"
                  :axios-params="tableQuery.axiosParams"
                />
              </v-col>
            </v-row>
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
              :loading="gridColumnMan.pending"
              :disabled="!isValid"
              >Save</v-btn
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
import { useActiveMetrics } from '@/metrics/use-metrics'
import { useGridColumnManager } from '@/metrics/use-dashboards'
import { useTableQuery } from '@/metrics/use-query'

// Components
import UnitPicker from '@/components/UnitPicker.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import TimeseriesTable from '@/metrics/TimeseriesTable.vue'

// Utilities
import { requiredRule } from '@/util/validation'
import { TableGridColumn } from '@/metrics/types'

export default defineComponent({
  name: 'DashGridColumnTableForm',
  components: {
    UnitPicker,
    MetricsPicker,
    MetricsQueryBuilder,
    TimeseriesTable,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    gridColumn: {
      type: Object as PropType<TableGridColumn>,
      required: true,
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const uql = useUql()

    const metricsPicker = shallowRef()
    const form = shallowRef()
    const isValid = shallowRef(false)
    const rules = { name: [requiredRule] }
    const gridColumnMan = useGridColumnManager()

    const activeMetrics = useActiveMetrics(computed(() => props.gridColumn.params.metrics))

    const tableQuery = useTableQuery(
      () => {
        if (!props.gridColumn.params.metrics.length || !props.gridColumn.params.query) {
          return undefined
        }

        return {
          ...props.dateRange.axiosParams(),
          metric: props.gridColumn.params.metrics.map((m) => m.name),
          alias: props.gridColumn.params.metrics.map((m) => m.alias),
          query: props.gridColumn.params.query,
        }
      },
      computed(() => props.gridColumn.params.columnMap),
    )

    watch(
      () => tableQuery.columns,
      () => {
        const unused = new Set(Object.keys(props.gridColumn.params.columnMap))

        for (let col of tableQuery.columns) {
          if (col.isGroup) {
            continue
          }
          unused.delete(col.name)
          if (col.name in props.gridColumn.params.columnMap) {
            continue
          }
          set(props.gridColumn.params.columnMap, col.name, {
            unit: col.unit,
          })
        }

        for (let colName of unused.values()) {
          del(props.gridColumn.params.columnMap, colName)
        }
      },
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
      () => tableQuery.query,
      (query) => {
        if (query) {
          uql.setQueryInfo(query)
        }
      },
      { immediate: true },
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

    return {
      uql,

      metricsPicker,
      form,
      isValid,
      rules,
      gridColumnMan,
      submit,

      activeMetrics,
      tableQuery,
    }
  },
})
</script>

<style lang="scss" scoped></style>
