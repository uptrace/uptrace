<template>
  <div>
    <v-menu v-model="menu" offset-y>
      <template #activator="{ on, attrs }">
        <v-btn color="primary" v-bind="attrs" v-on="on">
          <span>Add visualization</span>
          <v-icon right>mdi-menu-down</v-icon>
        </v-btn>
      </template>

      <v-list>
        <v-list-item @click="$emit('new', newChart())">
          <v-list-item-icon><v-icon>mdi-plus</v-icon></v-list-item-icon>
          <v-list-item-title>Chart</v-list-item-title>
        </v-list-item>
        <v-list-item @click="$emit('new', newTable())">
          <v-list-item-icon><v-icon>mdi-plus</v-icon></v-list-item-icon>
          <v-list-item-title>Table</v-list-item-title>
        </v-list-item>
        <v-list-item @click="$emit('new', newHeatmap())">
          <v-list-item-icon><v-icon>mdi-plus</v-icon></v-list-item-icon>
          <v-list-item-title>Heatmap</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, reactive, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Utilities
import {
  defaultChartLegend,
  Dashboard,
  ChartGridColumn,
  TableGridColumn,
  HeatmapGridColumn,
  GridColumnType,
  ChartKind,
} from '@/metrics/types'

export default defineComponent({
  name: 'DashGridColumnNewMenu',

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
  },

  setup(props, ctx) {
    const menu = shallowRef(false)

    function newChart(): ChartGridColumn {
      return reactive({
        id: 0,
        projectId: 0,
        dashId: 0,

        name: '',
        description: '',

        width: 0,
        height: 0,
        xAxis: 0,
        yAxis: 0,

        gridQueryTemplate: '',

        type: GridColumnType.Chart,
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

    function newTable(): TableGridColumn {
      return reactive({
        id: 0,
        projectId: 0,
        dashId: 0,

        name: '',
        description: '',

        width: 0,
        height: 0,
        xAxis: 0,
        yAxis: 0,

        gridQueryTemplate: '',

        type: GridColumnType.Table,
        params: {
          metrics: [],
          query: '',
          columnMap: {},
        },
      })
    }

    function newHeatmap(): HeatmapGridColumn {
      return reactive({
        id: 0,
        projectId: 0,
        dashId: 0,

        name: '',
        description: '',

        width: 0,
        height: 0,
        xAxis: 0,
        yAxis: 0,

        gridQueryTemplate: '',

        type: GridColumnType.Heatmap,
        params: {
          metric: '',
          unit: '',
          query: '',
        },
      })
    }

    return {
      GridColumnType,

      menu,

      newChart,
      newTable,
      newHeatmap,
    }
  },
})
</script>

<style lang="scss" scoped></style>
