<template>
  <div>
    <v-menu v-model="menu" offset-y>
      <template #activator="{ on, attrs }">
        <v-btn color="primary" small v-bind="attrs" v-on="on">
          <v-icon left>mdi-plus</v-icon>
          <span>Add</span>
          <v-icon right>mdi-menu-down</v-icon>
        </v-btn>
      </template>

      <v-list>
        <v-list-item v-if="dashKind === DashKind.Grid" @click="openDialog(newChart())">
          <v-list-item-title>Chart</v-list-item-title>
        </v-list-item>
        <v-list-item v-if="dashKind === DashKind.Grid" @click="openDialog(newTable())">
          <v-list-item-title>Table</v-list-item-title>
        </v-list-item>
        <v-list-item v-if="dashKind === DashKind.Grid" @click="openDialog(newHeatmap())">
          <v-list-item-title>Heatmap</v-list-item-title>
        </v-list-item>
        <v-list-item v-if="dashKind === DashKind.Grid" @click="addRow()">
          <v-list-item-title>Row</v-list-item-title>
        </v-list-item>
        <v-list-item @click="openDialog(newGauge())">
          <v-list-item-title>Gauge</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>

    <v-dialog v-model="dialog" fullscreen>
      <GridItemFormSwitch
        v-if="activeGridItem"
        :date-range="dateRange"
        :dashboard="dashboard"
        :table-grouping="dashboard.tableGrouping"
        :grid-item="activeGridItem"
        @save="
          dialog = false
          $emit('change')
        "
        @click:cancel="dialog = false"
      />
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, reactive, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useGridRowManager } from '@/metrics/use-dashboards'

// Components
import GridItemFormSwitch from '@/metrics/GridItemFormSwitch.vue'

// Misc
import {
  emptyBaseGridItem,
  defaultChartLegend,
  Dashboard,
  DashKind,
  GridItem,
  GridItemType,
  GaugeGridItem,
  ChartGridItem,
  TableGridItem,
  HeatmapGridItem,
  ChartKind,
} from '@/metrics/types'

export default defineComponent({
  name: 'NewGridItemMenu',
  components: { GridItemFormSwitch },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
    dashKind: {
      type: String as PropType<DashKind>,
      required: true,
    },
  },

  setup(props, ctx) {
    const menu = shallowRef(false)
    const gridRowMan = useGridRowManager()

    function addRow() {
      gridRowMan.create({ title: 'Row title', expanded: true }).then(() => {
        ctx.emit('change')
      })
    }

    function newGauge(): GaugeGridItem {
      return reactive({
        ...emptyBaseGridItem(),

        dashId: props.dashboard.id,
        dashKind: props.dashKind,

        type: GridItemType.Gauge,
        params: {
          metrics: [],
          query: '',
          columnMap: {},
          template: '',
          valueMappings: [],
        },
      })
    }

    function newChart(): ChartGridItem {
      return reactive({
        ...emptyBaseGridItem(),

        dashId: props.dashboard.id,
        dashKind: props.dashKind,

        type: GridItemType.Chart,
        params: {
          chartKind: ChartKind.Line,
          metrics: [],
          query: '',
          columnMap: {},
          timeseriesMap: {},
          legend: defaultChartLegend(),
        },
      })
    }

    function newTable(): TableGridItem {
      return reactive({
        ...emptyBaseGridItem(),

        dashId: props.dashboard.id,
        dashKind: props.dashKind,

        type: GridItemType.Table,
        params: {
          metrics: [],
          query: '',
          columnMap: {},
          itemsPerPage: 5,
          denseTable: false,
        },
      })
    }

    function newHeatmap(): HeatmapGridItem {
      return reactive({
        ...emptyBaseGridItem(),

        dashId: props.dashboard.id,
        dashKind: props.dashKind,

        type: GridItemType.Heatmap,
        params: {
          metric: '',
          unit: '',
          query: '',
        },
      })
    }

    const dialog = shallowRef(false)
    const activeGridItem = shallowRef<GridItem>()
    function openDialog(gridItem: GridItem) {
      activeGridItem.value = gridItem
      dialog.value = true
    }

    return {
      DashKind,

      menu,

      addRow,
      newGauge,
      newChart,
      newTable,
      newHeatmap,

      dialog,
      activeGridItem,
      openDialog,
    }
  },
})
</script>

<style lang="scss" scoped></style>
