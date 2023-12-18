<template>
  <GridItemFormPanes :grid-item="gridItem" v-on="$listeners">
    <template #picker>
      <MetricsPicker
        v-model="gridItem.params.metrics"
        :required-attrs="tableGrouping"
        :uql="uql"
        auto-grouping
      />
    </template>
    <template #preview>
      <v-container fluid>
        <v-row v-if="!activeMetrics.length">
          <v-col>
            <v-skeleton-loader type="table" height="400" boilerplate></v-skeleton-loader>
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
                <v-col cols="auto">
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
                </v-col>
              </v-row>

              <v-row>
                <v-col>
                  <GridItemTable
                    :date-range="dateRange"
                    :dashboard="dashboard"
                    :grid-item="gridItem"
                    @update:query="onTableQuery"
                    @update:columns="onTableColumns"
                  />
                </v-col>
              </v-row>
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

              <PanelSection title="Items per page">
                <v-slider
                  v-model="gridItem.params.itemsPerPage"
                  min="3"
                  max="20"
                  hide-details="auto"
                >
                  <template #prepend>{{ gridItem.params.itemsPerPage }}</template>
                </v-slider>
              </PanelSection>

              <v-checkbox v-model="gridItem.params.denseTable" label="Dense table" />
            </SinglePanel>
          </v-col>
        </v-row>
      </v-container>
    </template>
  </GridItemFormPanes>
</template>

<script lang="ts">
import { defineComponent, set, del, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql, BackendQueryInfo } from '@/use/uql'
import { useActiveMetrics } from '@/metrics/use-metrics'

// Components
import GridItemFormPanes from '@/metrics/GridItemFormPanes.vue'
import SinglePanel from '@/components/SinglePanel.vue'
import PanelSection from '@/components/PanelSection.vue'
import UnitPicker from '@/components/UnitPicker.vue'
import MetricsPicker from '@/metrics/MetricsPicker.vue'
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import GridItemTable from '@/metrics/GridItemTable.vue'

// Misc
import { requiredRule } from '@/util/validation'
import { Dashboard, TableGridItem, ColumnInfo } from '@/metrics/types'

export default defineComponent({
  name: 'GridItemTableForm',
  components: {
    GridItemFormPanes,
    SinglePanel,
    PanelSection,
    UnitPicker,
    MetricsPicker,
    MetricsQueryBuilder,
    GridItemTable,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
    tableGrouping: {
      type: Array as PropType<string[]>,
      required: true,
    },
    gridItem: {
      type: Object as PropType<TableGridItem>,
      required: true,
    },
  },

  setup(props, ctx) {
    const uql = useUql()

    const rules = { title: [requiredRule] }
    const activeMetrics = useActiveMetrics(computed(() => props.gridItem.params.metrics))

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

    function onTableQuery(query: BackendQueryInfo) {
      uql.setQueryInfo(query)
    }

    function onTableColumns(columns: ColumnInfo[]) {
      const unused = new Set(Object.keys(props.gridItem.params.columnMap))

      for (let col of columns) {
        if (col.isGroup) {
          continue
        }
        unused.delete(col.name)
        if (col.name in props.gridItem.params.columnMap) {
          continue
        }
        set(props.gridItem.params.columnMap, col.name, {
          unit: col.unit,
        })
      }

      for (let colName of unused.values()) {
        del(props.gridItem.params.columnMap, colName)
      }
    }

    return {
      uql,

      rules,

      activeMetrics,
      onTableColumns,
      onTableQuery,
    }
  },
})
</script>

<style lang="scss" scoped></style>
