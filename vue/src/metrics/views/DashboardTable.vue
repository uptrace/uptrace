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

    <v-row>
      <v-col>
        <v-card outlined rounded="lg">
          <v-toolbar flat color="blue lighten-5">
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

            <v-text-field
              v-model="tableQuery.searchInput"
              placeholder="Quick search: option1|option2"
              prepend-inner-icon="mdi-magnify"
              clearable
              outlined
              dense
              hide-details="auto"
              class="ml-8"
              style="max-width: 300px"
            />

            <v-spacer />

            <v-btn v-if="dashboard.tableQuery" class="primary" @click="dialog = true">
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
                  :columns="tableQuery.columns"
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

    <v-dialog v-model="dialog" max-width="1400">
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

    <v-dialog v-model="tableItem.dialog" max-width="1900">
      <DashTableDialogCard
        v-if="tableItem.dialog && tableItem.active"
        :date-range="dateRange"
        :dashboard="dashboard"
        :grid="grid"
        :grouping-columns="tableQuery.groupingColumns"
        :table-item="tableItem.active"
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
import { useUql, createQueryEditor } from '@/use/uql'
import { useActiveMetrics } from '@/metrics/use-metrics'
import { useTableQuery, TableItem } from '@/metrics/use-query'
import { useDashGauges } from '@/metrics/gauge/use-dash-gauges'

// Components
import ApiErrorCard from '@/components/ApiErrorCard.vue'
import GroupingToggle from '@/metrics/query/GroupingToggle.vue'
import TimeseriesTable from '@/metrics/TimeseriesTable.vue'
import DashTableForm from '@/metrics/DashTableForm.vue'
import DashTableDialogCard from '@/metrics/DashTableDialogCard.vue'
import DashGaugeRow from '@/metrics/gauge/DashGaugeRow.vue'

// Utilities
import { AttrKey } from '@/models/otel'
import { Dashboard, DashKind, GridColumn } from '@/metrics/types'

interface Props {
  dashboard: Dashboard
  grid: GridColumn[]
}

export default defineComponent({
  name: 'DashboardTable',
  components: {
    ApiErrorCard,
    GroupingToggle,
    TimeseriesTable,
    DashTableForm,
    DashTableDialogCard,
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

        const params = {
          ...props.dateRange.axiosParams(),
          time_offset: props.dashboard.timeOffset,
          metric: props.dashboard.tableMetrics.map((m) => m.name),
          alias: props.dashboard.tableMetrics.map((m) => m.alias),
          query: uql.query,
          min_interval: props.dashboard.minInterval,
        }
        return params
      },
      computed(() => props.dashboard.tableColumnMap),
    )

    useSyncQueryParams({
      fromQuery(queryParams) {
        queryParams.setDefault('query', props.dashboard.tableQuery)

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

    // Update the query when the dashboard is updated.
    watch(
      () => props.dashboard.tableQuery,
      (tableQuery) => {
        uql.query = tableQuery
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
        url: `/internal/v1/metrics/${projectId}/attr-keys`,
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

    const tableItem = useTableItem(props)

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

      cloneDeep,
      reactive,
    }
  },
})

function useTableItem(props: Props) {
  const dialog = shallowRef(false)
  const activeItem = shallowRef<TableItem>()

  const tableListeners = computed(() => {
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
