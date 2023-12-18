<template>
  <GridItemFormPanes :grid-item="gridItem" v-on="$listeners">
    <template #picker>
      <MetricsPicker v-model="gridItem.params.metrics" :required-attrs="tableGrouping" :uql="uql" />
    </template>
    <template #preview>
      <v-container fluid>
        <template v-if="!activeMetrics.length">
          <v-row>
            <v-col class="text-body-2">
              Text gauges are like <code>sprintf(format, values)</code>. You specify a
              <code>format</code> string with substitutions and Uptrace provides values. For
              example, using <code>${up_dbs} out of ${total_dbs} are up</code> format string you
              will get <code>5 out of 5 dbs are up</code> as the result.
            </v-col>
          </v-row>

          <v-row>
            <v-col>
              <v-skeleton-loader type="image" boilerplate></v-skeleton-loader>
            </v-col>
          </v-row>
        </template>

        <template v-else>
          <v-row>
            <v-col>
              <MetricsQueryBuilder
                :date-range="dateRange"
                :metrics="activeMetrics"
                :uql="uql"
                :disabled="!activeMetrics.length"
                show-agg
                show-dash-where
              />
            </v-col>
          </v-row>

          <v-row justify="center">
            <v-col cols="auto">
              <GaugeCard
                :loading="gaugeQuery.loading"
                :grid-item="gridItem"
                :columns="gaugeQuery.columns"
                :values="gaugeQuery.values"
                preview
              />
            </v-col>
          </v-row>
        </template>
      </v-container>
    </template>
    <template #options>
      <v-container fluid>
        <v-row>
          <v-col>
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
            </SinglePanel>
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <SinglePanel title="Gauge" expanded>
              <PanelSection title="Format string to customize gauge text">
                <v-text-field
                  v-model="gridItem.params.template"
                  placeholder="${num_db_up} dbs up out of ${num_db_total)}"
                  hint=""
                  persistent-hint
                  filled
                  dense
                  clearable
                  hide-details="auto"
                />
              </PanelSection>

              <GaugeValuesTable
                v-if="activeMetrics.length"
                :loading="gaugeQuery.loading"
                :grid-item="gridItem"
                :columns="gaugeQuery.columns"
                :values="gaugeQuery.values"
                :column-map="gridItem.params.columnMap"
              ></GaugeValuesTable>
            </SinglePanel>
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <SinglePanel title="Value mappings" expanded>
              <p>Use mappings to assign text and color to specific values, for example:</p>

              <ul class="mb-4">
                <li>0 &rarr; down (red)</li>
                <li>1 &rarr; up (green)</li>
              </ul>

              <v-dialog v-model="mappingsDialog" max-width="800">
                <template #activator="{ on, attrs }">
                  <v-btn block v-bind="attrs" v-on="on">Configure</v-btn>
                </template>

                <v-card>
                  <v-toolbar color="light-blue lighten-5" flat>
                    <v-toolbar-title>Value mappings</v-toolbar-title>
                  </v-toolbar>

                  <div class="pa-4">
                    <ValueMappingsForm
                      v-model="gridItem.params.valueMappings"
                      @click:close="mappingsDialog = false"
                    />
                  </div>
                </v-card>
              </v-dialog>
            </SinglePanel>
          </v-col>
        </v-row>
      </v-container>
    </template>
  </GridItemFormPanes>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import { useActiveMetrics } from '@/metrics/use-metrics'
import { formatGauge, useGaugeQuery } from '@/metrics/use-gauges'

// Components
import GridItemFormPanes from '@/metrics/GridItemFormPanes.vue'
import SinglePanel from '@/components/SinglePanel.vue'
import PanelSection from '@/components/PanelSection.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import GaugeCard from '@/metrics/GaugeCard.vue'
import GaugeValuesTable from '@/metrics/GaugeValuesTable.vue'
import ValueMappingsForm from '@/metrics/ValueMappingsForm.vue'

// Misc
import { requiredRule, minMaxStringLengthRule } from '@/util/validation'
import { updateColumnMap, GaugeGridItem } from '@/metrics/types'

export default defineComponent({
  name: 'GridItemGaugeForm',
  components: {
    GridItemFormPanes,
    SinglePanel,
    PanelSection,
    MetricsPicker,
    MetricsQueryBuilder,
    GaugeCard,
    GaugeValuesTable,
    ValueMappingsForm,
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
      type: Object as PropType<GaugeGridItem>,
      required: true,
    },
  },

  setup(props, ctx) {
    const mappingsDialog = shallowRef(false)
    const rules = { title: [requiredRule, minMaxStringLengthRule(0, 40)] }

    const uql = useUql()

    const activeMetrics = useActiveMetrics(computed(() => props.gridItem.params.metrics))

    const gaugeQuery = useGaugeQuery(
      () => {
        if (
          !props.gridItem ||
          !props.gridItem.params.metrics.length ||
          !props.gridItem.params.query
        ) {
          return undefined
        }

        return {
          ...props.dateRange.axiosParams(),
          metric: props.gridItem.params.metrics.map((m) => m.name),
          alias: props.gridItem.params.metrics.map((m) => m.alias),
          query: props.gridItem.params.query,
        }
      },
      computed(() => props.gridItem.params.columnMap),
    )

    const gaugeText = computed(() => {
      return formatGauge(
        gaugeQuery.values,
        gaugeQuery.columns,
        props.gridItem.params.template,
        'Add a metric first...',
      )
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
      () => gaugeQuery.query,
      (query) => {
        if (query) {
          uql.setQueryInfo(query)
        }
      },
    )
    watch(
      () => gaugeQuery.columns,
      (columns) => {
        updateColumnMap(props.gridItem.params.columnMap, columns)
      },
    )

    return {
      mappingsDialog,

      uql,

      activeMetrics,
      gaugeQuery,
      gaugeText,

      rules,
    }
  },
})
</script>

<style lang="scss" scoped></style>
