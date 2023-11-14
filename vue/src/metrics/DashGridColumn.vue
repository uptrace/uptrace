<template>
  <v-card flat>
    <v-toolbar color="light-blue lighten-5" flat dense>
      <v-toolbar-items>
        <v-btn
          icon
          title="Drag and drop to change position"
          class="drag-handle"
          style="cursor: move"
        >
          <v-icon>mdi-drag</v-icon>
        </v-btn>
      </v-toolbar-items>

      <v-tooltip v-if="columnError" bottom>
        <template #activator="{ on, attrs }">
          <v-toolbar-title class="d-flex align-center red--text" v-bind="attrs" v-on="on">
            <v-icon color="error" class="mr-2"> mdi-alert-circle-outline </v-icon>
            <span>{{ gridColumn.name }}</span>
          </v-toolbar-title>
        </template>
        <span>{{ columnError }}</span>
      </v-tooltip>
      <v-toolbar-title v-else class="d-flex align-center">
        {{ gridColumn.name }}
      </v-toolbar-title>

      <v-tooltip v-if="gridColumn.description" bottom>
        <template #activator="{ on, attrs }">
          <v-toolbar-items v-bind="attrs" v-on="on">
            <v-icon class="ml-2">mdi-information-outline</v-icon>
          </v-toolbar-items>
        </template>
        <span>{{ gridColumn.description }}</span>
      </v-tooltip>

      <v-spacer />

      <v-toolbar-items>
        <v-menu v-model="menu" offset-y>
          <template #activator="{ on: onMenu, attrs }">
            <v-btn :loading="gridColumnMan.pending" icon v-bind="attrs" v-on="onMenu">
              <v-icon>mdi-dots-vertical</v-icon>
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="dialog = true">
              <v-list-item-icon>
                <v-icon>mdi-eye</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>View</v-list-item-title>
              </v-list-item-content>
            </v-list-item>

            <v-list-item @click="$emit('click:edit', internalGridColumn)">
              <v-list-item-icon>
                <v-icon>{{ editable ? 'mdi-pencil' : 'mdi-lock' }}</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>Edit</v-list-item-title>
              </v-list-item-content>
            </v-list-item>

            <v-list-item v-if="routeForNewMonitor" :to="routeForNewMonitor">
              <v-list-item-icon>
                <v-icon>mdi-radar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>Monitor</v-list-item-title>
              </v-list-item-content>
            </v-list-item>

            <v-list-item v-if="editable" @click="del">
              <v-list-item-icon>
                <v-icon>mdi-delete</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>Delete</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
      </v-toolbar-items>
    </v-toolbar>

    <DashGridColumnItem
      :date-range="dateRange"
      :dashboard="dashboard"
      :grid-column="internalGridColumn"
      :height="height"
      @error="columnError = $event"
    />

    <v-dialog v-model="dialog" max-width="1200">
      <v-card outlined>
        <v-toolbar color="light-blue lighten-5" flat>
          <v-toolbar-title>{{ gridColumn.name }}</v-toolbar-title>
          <v-tooltip v-if="gridColumn.description" bottom>
            <template #activator="{ on, attrs }">
              <v-toolbar-items v-bind="attrs" v-on="on">
                <v-icon class="ml-2">mdi-information-outline</v-icon>
              </v-toolbar-items>
            </template>
            <span>{{ gridColumn.description }}</span>
          </v-tooltip>

          <v-spacer />

          <v-toolbar-items>
            <v-btn icon @click="dialog = false">
              <v-icon>mdi-close</v-icon>
            </v-btn>
          </v-toolbar-items>
        </v-toolbar>

        <div class="pa-2 pb-4">
          <DashGridColumnItem
            :date-range="dateRange"
            :dashboard="dashboard"
            :grid-column="internalGridColumn"
            :height="500"
            verbose
          />
        </div>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash-es'
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { Dashboard } from '@/metrics/types'
import { joinQuery } from '@/use/uql'
import { useGridColumnManager } from '@/metrics/use-dashboards'

// Components
import DashGridColumnItem from '@/metrics/DashGridColumnItem.vue'

// Utilities
import { GridColumn, GridColumnType } from '@/metrics/types'

export default defineComponent({
  name: 'DashGridColumn',
  components: {
    DashGridColumnItem,
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
    gridColumn: {
      type: Object as PropType<GridColumn>,
      required: true,
    },
    gridQuery: {
      type: String,
      required: true,
    },
    height: {
      type: Number,
      required: true,
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const menu = shallowRef(false)
    const dialog = shallowRef(false)

    const columnError = shallowRef(false)
    const routeForNewMonitor = computed(() => {
      switch (props.gridColumn.type) {
        case GridColumnType.Chart:
        case GridColumnType.Table:
          return {
            name: 'MonitorMetricNew',
            query: {
              name: props.gridColumn.name,
              metric: props.gridColumn.params.metrics.map((m) => m.name),
              alias: props.gridColumn.params.metrics.map((m) => m.alias),
              query: joinQuery([props.gridColumn.params.query, props.gridQuery]),
              columns: JSON.stringify(props.gridColumn.params.columnMap),
            },
          }
        default:
          return undefined
      }
    })

    const gridColumnMan = useGridColumnManager()

    const internalGridColumn = computed(() => {
      const gridColumn = cloneDeep(props.gridColumn)
      if (props.gridQuery) {
        gridColumn.params.query += ` | ${props.gridQuery}`
      }
      return gridColumn
    })

    function del() {
      gridColumnMan.del(props.gridColumn).then(() => {
        ctx.emit('change')
      })
    }

    return {
      GridColumnType,

      menu,
      dialog,

      columnError,
      routeForNewMonitor,

      gridColumnMan,
      internalGridColumn,
      del,
    }
  },
})
</script>

<style lang="scss" scoped></style>
