<template>
  <div>
    <v-row v-if="tableQuery.status.hasData()">
      <v-col>
        <v-sheet outlined rounded="lg" class="pa-2 px-4">
          <DashQueryBuilder :date-range="dateRange" :metrics="metricNames" :uql="baseUql">
            <template #prepend-actions>
              <v-btn
                v-if="tableQueryMan.isDirty"
                :loading="tableQueryMan.pending"
                small
                depressed
                class="mr-4"
                @click="tableQueryMan.save"
              >
                <v-icon small left>mdi-check</v-icon>
                <span>Save</span>
              </v-btn>
            </template>
          </DashQueryBuilder>
        </v-sheet>
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <v-card outlined rounded="lg">
          <v-toolbar flat color="blue lighten-5">
            <v-toolbar-title>
              <span :class="{ 'red--text': tableQuery.hasError }">{{ dashboard.data.name }}</span>
              <v-icon
                v-if="tableQuery.hasError"
                color="error"
                title="The query has errors"
                class="ml-2"
                >mdi-alert-circle-outline</v-icon
              >
            </v-toolbar-title>

            <v-spacer />

            <v-dialog v-model="dialog" max-width="1200">
              <template #activator="{ on, attrs }">
                <v-btn small outlined v-bind="attrs" v-on="on">Edit</v-btn>
              </template>

              <DashTableForm
                v-if="dialog"
                :date-range="dateRange"
                :metrics="metrics"
                :dashboard="dashboard"
                :table-query="tableQuery"
                :axios-params="axiosParams"
                @click:save="onSave"
                @click:cancel="onCancel"
              >
              </DashTableForm>
            </v-dialog>
          </v-toolbar>

          <v-card-text>
            <MetricItemsTable
              :loading="tableQuery.loading"
              :items="tableQuery.items"
              :columns="tableQuery.columns"
              :order="tableQuery.order"
              :axios-params="axiosParams"
              :column-map="dashboard.columnMap"
              v-on="itemViewer.listeners"
            >
            </MetricItemsTable>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <XPagination :pager="tableQuery.pager" />
      </v-col>
    </v-row>

    <v-dialog v-model="itemViewer.dialog" max-width="1400">
      <v-sheet v-if="itemViewer.dialog && itemViewer.active">
        <v-toolbar flat color="blue lighten-5">
          <v-toolbar-title
            >{{ dashboard.data.name }} {{ itemViewer.active[xkey.itemName] }}</v-toolbar-title
          >

          <v-spacer />

          <v-btn
            small
            outlined
            :loading="dashboard.loading"
            class="mr-4"
            @click="dashboard.reload()"
          >
            <v-icon small left>mdi-refresh</v-icon>
            <span>Reload</span>
          </v-btn>

          <v-btn icon @click="itemViewer.dialog = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar>

        <div class="pa-4">
          <DashGrid
            :date-range="dateRange"
            :metrics="metrics"
            :dashboard="dashboard"
            :base-query.sync="itemViewer.baseQuery"
          />
        </div>
      </v-sheet>
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, proxyRefs, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useTitle } from '@vueuse/core'
import { useUql } from '@/use/uql'
import { UseMetrics } from '@/metrics/use-metrics'
import { useDashQueryManager, UseDashboard } from '@/metrics/use-dashboards'
import { useTableQuery, TableItem } from '@/metrics/use-query'

// Components
import DashQueryBuilder from '@/metrics/query/DashQueryBuilder.vue'
import MetricItemsTable from '@/metrics/MetricItemsTable.vue'
import DashTableForm from '@/metrics/DashTableForm.vue'
import DashGrid from '@/metrics/DashGrid.vue'

// Utilities
import { xkey } from '@/models/otelattr'

interface Props {
  dashboard: UseDashboard
}

export default defineComponent({
  name: 'DashTable',
  components: {
    DashQueryBuilder,
    MetricItemsTable,
    DashTableForm,
    DashGrid,
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
    editing: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    useTitle(computed(() => `${props.dashboard.data?.name} | Metrics`))

    const dialog = shallowRef(false)
    const baseUql = useUql()
    const tableQueryMan = useDashQueryManager(props.dashboard)

    const metricNames = computed((): string[] | undefined => {
      if (!props.dashboard.status.isResolved()) {
        return []
      }
      return props.dashboard.metrics.map((m) => m.name)
    })

    const axiosParams = computed(() => {
      const dashData = props.dashboard.data
      if (!dashData || !dashData.query || !dashData.metrics || !dashData.metrics.length) {
        return { _: undefined }
      }

      return {
        ...props.dateRange.axiosParams(),
        metrics: dashData.metrics.map((m) => m.name),
        aliases: dashData.metrics.map((m) => m.alias),
        dash_id: dashData.id, // TODO: remove
        base_query: dashData.baseQuery,
        query: dashData.query,
      }
    })

    const tableQuery = useTableQuery(axiosParams, { syncQuery: true })

    watch(
      () => props.dashboard.data?.baseQuery ?? '',
      (query) => {
        baseUql.query = query
      },
      { immediate: true },
    )

    watch(
      () => baseUql.query,
      (query) => {
        if (props.dashboard.data) {
          props.dashboard.data.baseQuery = query
        }
      },
    )

    watch(
      () => tableQuery.baseQueryParts,
      (queryParts) => {
        if (queryParts) {
          baseUql.syncParts(queryParts)
        }
      },
      { immediate: true },
    )

    watch(
      () => props.editing,
      (editing) => {
        if (editing) {
          dialog.value = true
        }
      },
      { immediate: true },
    )

    function onSave() {
      props.dashboard.reload()
      dialog.value = false
      ctx.emit('change')
    }

    function onCancel() {
      props.dashboard.reload()
      dialog.value = false
      ctx.emit('change')
    }

    return {
      xkey,
      dialog,

      baseUql,
      metricNames,
      tableQuery,
      tableQueryMan,
      axiosParams,

      itemViewer: useItemViewer(props),

      onSave,
      onCancel,
    }
  },
})

function useItemViewer(props: Props) {
  const dialog = shallowRef(false)
  const activeItem = shallowRef<TableItem>()
  const baseQuery = shallowRef('')

  const listeners = computed(() => {
    if (props.dashboard.isTemplate && !props.dashboard.entries.length) {
      return {}
    }
    return {
      click: show,
    }
  })

  watch(activeItem, (item) => {
    baseQuery.value = item ? item[xkey.itemQuery] : ''
  })

  function show(item: TableItem) {
    activeItem.value = item
    dialog.value = true
  }

  return proxyRefs({ dialog, active: activeItem, baseQuery, listeners })
}
</script>

<style lang="scss" scoped></style>
