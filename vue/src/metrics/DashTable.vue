<template>
  <v-container :fluid="$vuetify.breakpoint.lgAndDown">
    <v-row v-if="dashGauges.items.length || editable" align="end" class="mb-n5">
      <v-col>
        <DashGaugeRow
          :date-range="dateRange"
          :dash-kind="DashKind.Table"
          :dash-gauges="dashGauges.items"
          :editable="editable"
          @change="dashGauges.reload"
        />
      </v-col>
    </v-row>

    <v-row v-if="tableQuery.status.hasData()">
      <v-col>
        <v-sheet outlined rounded="lg" class="pa-2 px-4">
          <MetricsQueryBuilder
            :date-range="dateRange"
            :metrics="activeMetrics"
            :uql="uql"
            show-agg
            show-group-by
            show-dash-where
            :disabled="!dashboard.tableMetrics.length"
          >
          </MetricsQueryBuilder>
        </v-sheet>
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <v-card outlined rounded="lg">
          <v-toolbar flat color="blue lighten-5">
            <v-tooltip v-if="tableQuery.error" bottom>
              <template #activator="{ on, attrs }">
                <v-toolbar-items v-bind="attrs" v-on="on">
                  <v-icon color="error" class="mr-2"> mdi-alert-circle-outline </v-icon>
                </v-toolbar-items>
              </template>
              <span>{{ tableQuery.error }}</span>
            </v-tooltip>

            <v-toolbar-title :class="{ 'red--text': Boolean(tableQuery.error) }">
              {{ dashboard.name }}
            </v-toolbar-title>

            <v-text-field
              v-model="tableQuery.searchInput"
              placeholder="Quick search"
              clearable
              outlined
              dense
              hide-details="auto"
              class="ml-8"
              style="max-width: 300px"
            />

            <v-spacer />

            <v-dialog v-model="dialog" max-width="1400">
              <template #activator="{ on, attrs }">
                <v-btn class="primary" v-bind="attrs" v-on="on">
                  <v-icon left>mdi-pencil</v-icon>
                  <span>Edit table</span>
                </v-btn>
              </template>

              <DashTableForm
                v-if="dialog"
                :date-range="dateRange"
                :dashboard="reactive(cloneDeep(dashboard))"
                :editable="editable"
                @click:save="
                  dialog = false
                  $emit('change')
                "
                @click:cancel="dialog = false"
              >
              </DashTableForm>
            </v-dialog>
          </v-toolbar>

          <v-container fluid>
            <v-row v-if="tableQuery.items.length" align="center" justify="space-between">
              <v-col cols="auto">
                <GroupingToggle
                  v-if="attrKeysDs.items.length"
                  v-model="grouping"
                  :loading="tableQuery.loading || attrKeysDs.loading"
                  :items="attrKeysDs.items"
                />
              </v-col>
              <v-col v-if="attrKeysDs.items.length < 5" cols="auto" class="text--secondary">
                Click on a row to view the Grid filtered by <code>group by</code> attributes.
              </v-col>
            </v-row>

            <v-row>
              <v-col>
                <TimeseriesTable
                  :loading="tableQuery.loading"
                  :items="tableQuery.items"
                  :columns="tableQuery.columns"
                  :order="tableQuery.order"
                  :axios-params="tableQuery.axiosParams"
                  v-on="tableItem.listeners"
                >
                </TimeseriesTable>
              </v-col>
            </v-row>
          </v-container>
        </v-card>
      </v-col>
    </v-row>

    <v-dialog v-model="tableItem.dialog" max-width="1900">
      <v-sheet v-if="tableItem.dialog && tableItem.active">
        <v-toolbar flat color="blue lighten-5">
          <v-toolbar-title>{{ dashboard.name }} {{ tableItem.active._name }}</v-toolbar-title>

          <v-spacer />

          <v-btn small outlined class="mr-4" @click="dateRange.reload">
            <v-icon small left>mdi-refresh</v-icon>
            <span>Reload</span>
          </v-btn>

          <v-toolbar-items>
            <v-btn icon @click="tableItem.dialog = false">
              <v-icon>mdi-close</v-icon>
            </v-btn>
          </v-toolbar-items>
        </v-toolbar>

        <DashGrid
          :date-range="dateRange"
          :dashboard="dashboard"
          :grid="grid"
          :grid-query="gridQueryFor(tableItem.active)"
          :grouping-columns="tableQuery.groupingColumns"
          :table-item="tableItem.active"
        />
      </v-sheet>
    </v-dialog>
  </v-container>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash-es'
import { defineComponent, shallowRef, reactive, computed, proxyRefs, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useTitle } from '@vueuse/core'
import { useRoute } from '@/use/router'
import { useDataSource } from '@/use/datasource'
import { useUql, createUqlEditor } from '@/use/uql'
import { useActiveMetrics } from '@/metrics/use-metrics'
import { useTableQuery, TableItem } from '@/metrics/use-query'
import { useDashGauges } from '@/metrics/gauge/use-dash-gauges'

// Components
import MetricsQueryBuilder from '@/metrics/query/MetricsQueryBuilder.vue'
import GroupingToggle from '@/metrics/query/GroupingToggle.vue'
import TimeseriesTable from '@/metrics/TimeseriesTable.vue'
import DashTableForm from '@/metrics/DashTableForm.vue'
import DashGrid from '@/metrics/DashGrid.vue'
import DashGaugeRow from '@/metrics/gauge/DashGaugeRow.vue'

// Utilities
import { AttrKey } from '@/models/otel'
import { Dashboard, DashKind, GridColumn } from '@/metrics/types'

interface Props {
  dashboard: Dashboard
  grid: GridColumn[]
}

export default defineComponent({
  name: 'DashTable',
  components: {
    MetricsQueryBuilder,
    GroupingToggle,
    TimeseriesTable,
    DashTableForm,
    DashGrid,
    DashGaugeRow,
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
    grid: {
      type: Array as PropType<GridColumn[]>,
      required: true,
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    useTitle(computed(() => `${props.dashboard.name} | Metrics`))

    const route = useRoute()
    const dialog = shallowRef(false)
    const uql = useUql()

    const dashGauges = useDashGauges(() => {
      return {
        dash_kind: DashKind.Table,
      }
    })

    const activeMetrics = useActiveMetrics(computed(() => props.dashboard.tableMetrics))

    const tableQuery = useTableQuery(
      () => {
        if (!props.dashboard.tableQuery || !props.dashboard.tableMetrics.length) {
          return { _: undefined }
        }

        return {
          ...props.dateRange.axiosParams(),
          metric: props.dashboard.tableMetrics.map((m) => m.name),
          alias: props.dashboard.tableMetrics.map((m) => m.alias),
          query: uql.query,
        }
      },
      computed(() => props.dashboard.tableColumnMap),
    )
    tableQuery.order.syncQueryParams()

    watch(
      () => props.dashboard.tableQuery ?? '',
      (query) => {
        uql.query = query
      },
      { immediate: true },
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

    const attrKeysDs = useDataSource(() => {
      if (!props.dashboard.tableMetrics.length) {
        return undefined
      }

      const { projectId } = route.value.params
      return {
        url: `/api/v1/metrics/${projectId}/attr-keys`,
        params: {
          ...props.dateRange.axiosParams(),
          metric: props.dashboard.tableMetrics.map((m) => m.name),
        },
      }
    })

    const grouping = computed({
      get() {
        return tableQuery.columns.filter((col) => col.isGroup).map((column) => column.name)
      },
      set(grouping: string[]) {
        const editor = createUqlEditor()

        for (let colName of grouping) {
          editor.groupBy(colName)
        }

        for (let part of uql.parts) {
          if (/^group by/i.test(part.query)) {
            continue
          }
          editor.add(part.query)
        }

        uql.commitEdits(editor)
      },
    })

    const tableItem = useTableItem(props)
    function gridQueryFor(tableItem: TableItem): string {
      const ss = []

      if (tableItem._query) {
        ss.push(tableItem._query)
      }

      if (props.dashboard.gridQuery) {
        ss.push(props.dashboard.gridQuery)
      }

      return ss.join(' | ')
    }

    return {
      AttrKey,
      DashKind,

      dialog,
      dashGauges,

      uql,
      activeMetrics,
      tableQuery,

      attrKeysDs,
      grouping,

      tableItem,
      gridQueryFor,

      cloneDeep,
      reactive,
    }
  },
})

function useTableItem(props: Props) {
  const dialog = shallowRef(false)
  const activeItem = shallowRef<TableItem>()

  const tableListeners = computed(() => {
    if (!props.grid.length) {
      return {}
    }
    return {
      click(item: TableItem) {
        activeItem.value = item
        dialog.value = true
      },
    }
  })

  return proxyRefs({
    dialog,
    active: activeItem,
    listeners: tableListeners,
  })
}
</script>

<style lang="scss" scoped></style>
