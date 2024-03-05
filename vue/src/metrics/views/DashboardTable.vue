<template>
  <v-container :fluid="$vuetify.breakpoint.lgAndDown">
    <portal to="dashboard-actions">
      <v-col cols="auto">
        <NewGridItemMenu
          :date-range="dateRange"
          :dashboard="dashboard"
          :dash-kind="DashKind.Table"
          @change="$emit('change', $event)"
        />
      </v-col>
    </portal>

    <v-row v-if="dashboard.tableMetrics.length" dense>
      <v-col>
        <v-card outlined rounded="lg" class="py-2 px-4">
          <DashQueryBuilder
            :date-range="dateRange"
            :metrics="dashboard.tableMetrics.map((m) => m.name)"
            :uql="uql"
          >
          </DashQueryBuilder>
        </v-card>
      </v-col>
    </v-row>

    <v-row v-if="tableItems.length" dense>
      <v-col>
        <GridStackCard :items="tableItems">
          <template #item="{ attrs, on }">
            <GridItemAny
              :date-range="dateRange"
              :dashboard="dashboard"
              v-bind="attrs"
              v-on="{
                ...on,
                change() {
                  $emit('change')
                },
              }"
            />
          </template>
        </GridStackCard>
      </v-col>
    </v-row>

    <v-row dense>
      <v-col>
        <v-card outlined rounded="lg">
          <v-toolbar flat dense color="blue lighten-5">
            <v-tooltip v-if="tableQuery.queryError" bottom>
              <template #activator="{ on, attrs }">
                <v-toolbar-items v-bind="attrs" v-on="on">
                  <v-icon color="error" class="mr-2">mdi-alert-circle-outline</v-icon>
                </v-toolbar-items>
              </template>
              <span>{{ tableQuery.queryError }}</span>
            </v-tooltip>

            <v-toolbar-title :class="{ 'red--text': Boolean(tableQuery.queryError) }">
              {{ dashboard.name }}
            </v-toolbar-title>

            <QuickSearch v-model="tableQuery.searchInput" class="ml-8" />

            <v-spacer />

            <v-btn v-if="dashboard.tableQuery" small class="primary" @click="dialog = true">
              <v-icon left>mdi-pencil</v-icon>
              <span>Edit table</span>
            </v-btn>
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

            <v-row v-if="dashboard.tableQuery">
              <v-col>
                <ApiErrorCard v-if="tableQuery.error" :error="tableQuery.error" />
                <TimeseriesTable
                  v-else
                  :loading="tableQuery.loading"
                  :items="tableQuery.items"
                  :agg-columns="tableQuery.aggColumns"
                  :grouping-columns="tableQuery.groupingColumns"
                  :order="tableQuery.order"
                  :axios-params="tableQuery.axiosParams"
                  v-on="tableItem.listeners"
                >
                </TimeseriesTable>
              </v-col>
            </v-row>
            <v-row v-else class="py-10">
              <v-col>
                <v-row>
                  <v-col class="text-center">
                    This dashboard is empty. To add some metrics, click on the "Edit table"
                    button.<br />
                    If you are not familiar with Uptrace dashboards, click on the "Help" button.
                  </v-col>
                </v-row>

                <v-row justify="center">
                  <v-col cols="auto">
                    <v-btn class="primary" @click="dialog = true">
                      <v-icon left>mdi-pencil</v-icon>
                      <span>Edit table</span>
                    </v-btn>
                  </v-col>
                  <v-col cols="auto">
                    <v-btn :to="{ name: 'DashboardHelp' }">
                      <v-icon left>mdi-help-circle-outline</v-icon>
                      <span>Help</span>
                    </v-btn>
                  </v-col>
                </v-row>
              </v-col>
            </v-row>
          </v-container>
        </v-card>
      </v-col>
    </v-row>

    <DashTableFormDialog
      v-model="dialog"
      :date-range="dateRange"
      :dashboard="reactive(cloneDeep(dashboard))"
      @saved="$emit('change')"
    >
    </DashTableFormDialog>

    <v-dialog v-model="tableItem.dialog" fullscreen>
      <DashGridForTableRow
        v-if="tableItem.dialog && tableItem.active"
        :date-range="dateRange"
        :dashboard="dashboard"
        :grid-rows="gridRows"
        :grid-metrics="gridMetrics"
        :table-row="tableItem.active"
        :table-grouping="tableQuery.groupingColumns"
        @click:close="tableItem.dialog = false"
      />
    </v-dialog>
  </v-container>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash-es'
import { defineComponent, shallowRef, reactive, computed, proxyRefs, watch, PropType } from 'vue'

// Composables
import { useSyncQueryParams } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useRoute } from '@/use/router'
import { useDataSource } from '@/use/datasource'
import { useUql, joinQuery, createQueryEditor } from '@/use/uql'
import { useTableQuery } from '@/metrics/use-query'

// Components
import DashQueryBuilder from '@/metrics/query/DashQueryBuilder.vue'
import NewGridItemMenu from '@/metrics/NewGridItemMenu.vue'
import GroupingToggle from '@/metrics/query/GroupingToggle.vue'
import TimeseriesTable from '@/metrics/TimeseriesTable.vue'
import DashTableFormDialog from '@/metrics/DashTableFormDialog.vue'
import DashGridForTableRow from '@/metrics/DashGridForTableRow.vue'
import ApiErrorCard from '@/components/ApiErrorCard.vue'
import GridStackCard from '@/metrics/GridStackCard.vue'
import GridItemAny from '@/metrics/GridItemAny.vue'
import QuickSearch from '@/components/QuickSearch.vue'

// Misc
import { AttrKey } from '@/models/otel'
import { Dashboard, DashKind, GridRow, GridItem, TableRowData } from '@/metrics/types'

export default defineComponent({
  name: 'DashboardTable',
  components: {
    DashQueryBuilder,
    NewGridItemMenu,
    GroupingToggle,
    TimeseriesTable,
    DashTableFormDialog,
    DashGridForTableRow,
    ApiErrorCard,
    GridStackCard,
    GridItemAny,
    QuickSearch,
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
    tableItems: {
      type: Array as PropType<GridItem[]>,
      required: true,
    },
    gridRows: {
      type: Array as PropType<GridRow[]>,
      required: true,
    },
    gridMetrics: {
      type: Array as PropType<string[]>,
      required: true,
    },
  },

  setup(props, ctx) {
    const route = useRoute()
    const dialog = shallowRef(false)

    const re = /^(where|group\s+by)\s+/i
    const baseQuery = computed(() => {
      return createQueryEditor(props.dashboard.tableQuery)
        .filter((part) => !re.test(part))
        .toString()
    })
    const editableQuery = computed(() => {
      return createQueryEditor(props.dashboard.tableQuery)
        .filter((part) => re.test(part))
        .toString()
    })

    const uql = useUql()
    watch(
      editableQuery,
      (query) => {
        uql.query = query
      },
      { immediate: true },
    )

    const tableQuery = useTableQuery(
      () => {
        if (!props.dashboard.tableQuery || !props.dashboard.tableMetrics.length) {
          return { _: undefined }
        }

        return {
          ...props.dateRange.axiosParams(),
          time_offset: props.dashboard.timeOffset,
          metric: props.dashboard.tableMetrics.map((m) => m.name),
          alias: props.dashboard.tableMetrics.map((m) => m.alias),
          query: joinQuery([uql.query, baseQuery.value]),
          min_interval: props.dashboard.minInterval,
        }
      },
      computed(() => props.dashboard.tableColumnMap),
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

    useSyncQueryParams({
      fromQuery(queryParams) {
        queryParams.setDefault('query', editableQuery.value)

        props.dateRange.parseQueryParams(queryParams)
        tableQuery.order.parseQueryParams(queryParams)
        uql.parseQueryParams(queryParams)
      },
      toQuery() {
        return {
          ...props.dateRange.queryParams(),
          ...tableQuery.order.queryParams(),
          ...uql.queryParams(),
        }
      },
    })

    const attrKeysDs = useDataSource(() => {
      if (!props.dashboard.tableMetrics.length) {
        return undefined
      }

      const { projectId } = route.value.params
      return {
        url: `/internal/v1/metrics/${projectId}/attributes`,
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
        const editor = createQueryEditor()

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

    const tableItem = useTableRowData()

    return {
      AttrKey,
      DashKind,

      dialog,

      uql,
      tableQuery,

      attrKeysDs,
      grouping,

      tableItem,

      cloneDeep,
      reactive,
    }
  },
})

function useTableRowData() {
  const dialog = shallowRef(false)
  const activeItem = shallowRef<TableRowData>()

  const tableListeners = computed(() => {
    return {
      click(item: TableRowData) {
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
