<template>
  <v-sheet>
    <v-toolbar flat color="blue lighten-5">
      <v-toolbar-title>{{ dashboard.name }} {{ tableItem._name }}</v-toolbar-title>

      <v-spacer />

      <v-col cols="auto">
        <DateRangePicker :date-range="internalDateRange" :range-days="90" />
      </v-col>

      <v-toolbar-items>
        <v-btn icon @click="$emit('click:close')">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-toolbar-items>
    </v-toolbar>

    <DashGrid
      v-if="grid.length"
      :date-range="internalDateRange"
      :dashboard="dashboard"
      :grid="grid"
      :grid-query="gridQueryFor(tableItem)"
      :grouping-columns="groupingColumns"
      :table-item="tableItem"
    />
    <v-container v-else class="py-16" style="max-width: 800px">
      <v-row>
        <v-col>
          This grid dashboard does not contain any charts. You can add some in the
          <router-link :to="{ name: 'DashboardGrid' }">Grid</router-link> tab.
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          Grid dashboards are used together with table dashboards. Each row in the table dashboard
          leads to the same grid dashboard filtered by <code>group by</code> attributes from the
          table row, for example,
          <code>where host.name = ${host} and service.name = ${service}.</code>.
        </v-col>
      </v-row>

      <v-row justify="center">
        <v-col cols="auto">
          <v-btn :to="{ name: 'DashboardGrid' }" color="primary"> Configure grid </v-btn>
        </v-col>
        <v-col cols="auto">
          <v-btn :to="{ name: 'DashboardHelp' }"> Learn more </v-btn>
        </v-col>
      </v-row>
    </v-container>
  </v-sheet>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useDateRangeFrom, UseDateRange } from '@/use/date-range'
import { TableItem } from '@/metrics/use-query'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import DashGrid from '@/metrics/DashGrid.vue'

// Utilities
import { Dashboard, GridColumn } from '@/metrics/types'

export default defineComponent({
  name: 'DashTableDialogCard',
  components: { DateRangePicker, DashGrid },

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
    groupingColumns: {
      type: Array,
      required: true,
    },
    tableItem: {
      type: Object as PropType<TableItem>,
      required: true,
    },
  },

  setup(props) {
    const internalDateRange = useDateRangeFrom(props.dateRange)

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

    return { internalDateRange, gridQueryFor }
  },
})
</script>

<style lang="scss" scoped></style>
