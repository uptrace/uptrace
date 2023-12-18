<template>
  <v-card color="grey lighten-5" min-height="100vh">
    <v-card flat color="blue lighten-5">
      <v-card flat color="transparent" :max-width="dashboard.gridMaxWidth" class="mx-auto">
        <v-toolbar flat color="transparent">
          <v-toolbar-items>
            <v-btn icon @click="$emit('click:close')">
              <v-icon>mdi-close</v-icon>
            </v-btn>
          </v-toolbar-items>
          <v-toolbar-title>{{ dashboard.name }}</v-toolbar-title>

          <v-spacer />

          <v-col cols="auto">
            <DateRangePicker :date-range="internalDateRange" :range-days="90" />
          </v-col>
        </v-toolbar>
      </v-card>
    </v-card>

    <DashGrid
      v-if="gridRows.length"
      :date-range="internalDateRange"
      :dashboard="dashboard"
      :grid-rows="gridRows"
      :grid-metrics="gridMetrics"
      :grid-query="gridQuery"
      :table-row="tableRow"
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
  </v-card>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useDateRangeFrom, UseDateRange } from '@/use/date-range'
import { joinQuery } from '@/use/uql'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import DashGrid from '@/metrics/DashGrid.vue'

// Misc
import { Dashboard, GridRow, TableRowData } from '@/metrics/types'

export default defineComponent({
  name: 'DashGridForTableRow',
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
    gridRows: {
      type: Array as PropType<GridRow[]>,
      required: true,
    },
    gridMetrics: {
      type: Array as PropType<string[]>,
      required: true,
    },
    tableRow: {
      type: Object as PropType<TableRowData>,
      required: true,
    },
    maxWidth: {
      type: [Number, String],
      default: 1900,
    },
  },

  setup(props) {
    const internalDateRange = useDateRangeFrom(props.dateRange)

    const gridQuery = computed(() => {
      const ss = []
      if (props.dashboard.gridQuery) {
        ss.push(props.dashboard.gridQuery)
      }
      if (props.tableRow._query) {
        ss.push(props.tableRow._query)
      }
      return joinQuery(ss)
    })

    return { internalDateRange, gridQuery }
  },
})
</script>

<style lang="scss" scoped></style>
