<template>
  <v-container style="max-width: 1900px">
    <v-row v-if="grid.length">
      <v-col>
        <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="pa-0">
          <v-card outlined rounded="lg" class="py-2 px-4">
            <DashQueryBuilder
              :date-range="dateRange"
              :metrics="metricNames"
              :uql="uql"
              class="mb-1"
            >
              <template #prepend-actions>
                <v-btn
                  v-if="!tableItem && isGridQueryDirty"
                  :loading="dashMan.pending"
                  small
                  depressed
                  class="mr-4"
                  @click="saveGridQuery"
                >
                  <v-icon small left>mdi-check</v-icon>
                  <span>Save</span>
                </v-btn>
              </template>
            </DashQueryBuilder>
          </v-card>
        </v-container>
      </v-col>
    </v-row>

    <v-row align="end">
      <v-col v-if="editable" cols="auto">
        <DashGridColumnNewMenu
          :date-range="dateRange"
          :dashboard="dashboard"
          @new="
            activeGridColumn = $event
            dialog = true
          "
        />
      </v-col>
      <v-col v-if="dashGauges.items.length || editable">
        <DashGaugeRow
          :date-range="dateRange"
          :dash-kind="DashKind.Grid"
          :grid-query="uql.query"
          :editable="editable"
          :dash-gauges="dashGauges.items"
          @change="dashGauges.reload"
        />
      </v-col>
    </v-row>

    <v-row v-if="!internalGrid.length">
      <v-col v-for="i in 6" :key="i" cols="6">
        <v-skeleton-loader type="image" boilerplate></v-skeleton-loader>
      </v-col>
    </v-row>

    <v-row v-else>
      <v-col>
        <div ref="gridStackEl" class="grid-stack">
          <div
            v-for="gridColumn in internalGrid"
            :id="`gsi-${gridColumn.id}`"
            :key="gridColumn.id"
            :gs-id="gridColumn.id"
            :gs-w="gridColumn.width"
            :gs-h="gridColumn.height"
            :gs-x="gridColumn.xAxis"
            :gs-y="gridColumn.yAxis"
            :gs-auto-position="gridAutoPosition"
            class="grid-stack-item"
          >
            <div class="grid-stack-item-content">
              <DashGridColumn
                :date-range="dateRange"
                :dashboard="dashboard"
                :grid-column="gridColumn"
                :grid-query="gridQueryFor(gridColumn)"
                :height="gridColumn.height * gridCellHeight - 64"
                :editable="editable"
                @change="$emit('change')"
                @click:edit="
                  activeGridColumn = $event
                  dialog = true
                "
              />
            </div>
          </div>
        </div>
      </v-col>
    </v-row>

    <v-dialog v-model="dialog" max-width="1200">
      <v-skeleton-loader v-if="!activeGridColumn" type="card"></v-skeleton-loader>
      <DashGridColumnChartForm
        v-else-if="activeGridColumn.type === GridColumnType.Chart"
        :date-range="dateRange"
        :table-grouping="dashboard.tableGrouping"
        :grid-column="activeGridColumn"
        :editable="editable"
        @click:save="
          dialog = false
          $emit('change')
        "
        @click:cancel="dialog = false"
      />
      <DashGridColumnTableForm
        v-else-if="activeGridColumn.type === GridColumnType.Table"
        :date-range="dateRange"
        :grid-column="activeGridColumn"
        :editable="editable"
        @click:save="
          dialog = false
          $emit('change')
        "
        @click:cancel="dialog = false"
      />
      <DashGridColumnHeatmapForm
        v-else-if="activeGridColumn.type === GridColumnType.Heatmap"
        :date-range="dateRange"
        :grid-column="activeGridColumn"
        :editable="editable"
        @click:save="
          dialog = false
          $emit('change')
        "
        @click:cancel="dialog = false"
      />
    </v-dialog>
  </v-container>
</template>

<script lang="ts">
import {
  defineComponent,
  shallowRef,
  ref,
  computed,
  watch,
  onMounted,
  onBeforeUnmount,
  nextTick,
  PropType,
} from 'vue'

import 'gridstack/dist/gridstack.min.css'
import 'gridstack/dist/gridstack-extra.min.css'
import { GridStack, GridStackOptions } from 'gridstack'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useTitle } from '@vueuse/core'
import { useUql } from '@/use/uql'
import { useDashManager, useGridColumnManager } from '@/metrics/use-dashboards'
import { useDashGauges } from '@/metrics/gauge/use-dash-gauges'
import { TableItem } from '@/metrics/use-query'

// Components
import DashQueryBuilder from '@/metrics/query/DashQueryBuilder.vue'
import DashGridColumn from '@/metrics/DashGridColumn.vue'
import DashGaugeRow from '@/metrics/gauge/DashGaugeRow.vue'
import DashGridColumnNewMenu from '@/metrics/DashGridColumnNewMenu.vue'
import DashGridColumnChartForm from '@/metrics/DashGridColumnChartForm.vue'
import DashGridColumnTableForm from '@/metrics/DashGridColumnTableForm.vue'
import DashGridColumnHeatmapForm from '@/metrics/DashGridColumnHeatmapForm.vue'

// Utilities
import { Dashboard, GridColumn, GridColumnType, DashKind } from '@/metrics/types'
import { quote } from '@/util/string'

export default defineComponent({
  name: 'DashGrid',
  components: {
    DashQueryBuilder,
    DashGridColumn,
    DashGaugeRow,
    DashGridColumnNewMenu,
    DashGridColumnChartForm,
    DashGridColumnTableForm,
    DashGridColumnHeatmapForm,
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
    gridQuery: {
      type: String,
      default: '',
    },
    groupingColumns: {
      type: Array as PropType<string[]>,
      default: undefined,
    },
    tableItem: {
      type: Object as PropType<TableItem>,
      default: undefined,
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    useTitle(computed(() => `${props.dashboard.name} | Metrics`))

    const uql = useUql()
    watch(
      () => props.gridQuery,
      (query) => {
        uql.query = query
      },
      { immediate: true },
    )

    const dashMan = useDashManager()
    const isGridQueryDirty = computed(() => {
      return uql.query !== props.gridQuery
    })
    function saveGridQuery() {
      dashMan.update({ gridQuery: uql.query }).then(() => {
        ctx.emit('change')
      })
    }

    const internalGrid = ref<GridColumn[]>([])
    watch(
      () => props.grid,
      (grid) => {
        internalGrid.value = grid
      },
      { immediate: true },
    )

    const gridCellHeight = 20
    const gridAutoPosition = computed(() => {
      return props.grid.every((col) => col.xAxis === 0 && col.yAxis === 0)
    })
    const gridStackEl = shallowRef()
    let gridStack: GridStack | undefined

    onMounted(() => {
      watch(
        () => props.grid,
        (grid) => {
          if (gridStack) {
            gridStack.destroy(false)
          }
          if (!grid.length) {
            return
          }
          nextTick(() => {
            gridStack = initGridStack()
          })
        },
        { immediate: true },
      )
    })
    onBeforeUnmount(() => {
      if (gridStack) {
        gridStack.destroy(false)
        gridStack = undefined
      }
    })

    function initGridStack() {
      const options: GridStackOptions = {
        oneColumnSize: 1000,
        cellHeight: gridCellHeight,
        margin: 5,
        minRow: 5,
        draggable: {
          handle: '.drag-handle',
        },
        resizable: { handles: 'se,sw' },
        //float: true,
      }

      const gridStack = GridStack.init(options, gridStackEl.value)

      gridStack.on('dragstop', updateGridPos)
      gridStack.on('resizestop', updateGridPos)

      return gridStack
    }

    function updateGridPos() {
      if (!gridStack || gridStack.getColumn() <= 1) {
        return
      }

      const data = []

      const items = gridStack.getGridItems()
      for (let el of items) {
        const node = el.gridstackNode
        if (!node) {
          continue
        }

        const id = typeof node.id === 'string' ? parseInt(node.id, 10) : node.id
        if (!id) {
          continue
        }

        data.push({
          id,
          width: node.w || 0,
          height: node.h || 0,
          xAxis: node.x || 0,
          yAxis: node.y || 0,
        })

        const cell = internalGrid.value.find((cell) => cell.id === id)
        if (cell) {
          cell.width = node.w || cell.width
          cell.height = node.h || cell.height
        }
      }

      gridColumnMan.updateOrder(data)
    }

    const internalDialog = shallowRef(false)
    const activeGridColumn = ref<GridColumn>()
    const dialog = computed({
      get(): boolean {
        return Boolean(internalDialog.value && activeGridColumn.value)
      },
      set(dialog: boolean) {
        internalDialog.value = dialog
      },
    })

    const gridColumnMan = useGridColumnManager()

    const metricNames = computed((): string[] => {
      const names: string[] = []
      for (let gridCol of props.grid) {
        switch (gridCol.type) {
          case GridColumnType.Chart:
          case GridColumnType.Table:
            for (let m of gridCol.params.metrics) {
              names.push(m.name)
            }
            break
          case GridColumnType.Heatmap:
            names.push(gridCol.params.metric)
            break
        }
      }
      return names
    })

    const dashGauges = useDashGauges(() => {
      return {
        dash_kind: DashKind.Grid,
      }
    })

    watch(dialog, (dialog) => {
      if (!dialog) {
        activeGridColumn.value = undefined
      }
    })

    function gridQueryFor(gridColumn: GridColumn) {
      if (props.groupingColumns && props.tableItem && gridColumn.gridQueryTemplate) {
        let gridQuery = gridColumn.gridQueryTemplate
        for (let colName of props.groupingColumns) {
          const varName = '${' + colName + '}'
          const varValue = quote(props.tableItem[colName])
          gridQuery = gridQuery.replaceAll(varName, String(varValue))
        }
        return gridQuery
      }
      return uql.query
    }

    return {
      DashKind,
      GridColumnType,

      uql,

      internalGrid,
      gridAutoPosition,
      gridCellHeight,
      gridStackEl,

      dashMan,
      isGridQueryDirty,
      saveGridQuery,

      gridColumnMan,
      activeGridColumn,
      dialog,
      gridQueryFor,

      metricNames,

      dashGauges,
    }
  },
})
</script>

<style lang="scss" scoped>
.grid-stack {
}

.grid-stack-item-content {
  border: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
