<template>
  <div>
    <v-simple-table class="v-data-table--large">
      <thead class="v-data-table-header">
        <tr>
          <ThOrder v-if="column" :value="column" :order="order">{{ column }}</ThOrder>
          <ThOrder value="system" :order="order">System</ThOrder>
          <ThOrder value="rate" :order="order" align="center">Spans per minute</ThOrder>
          <ThOrder value="errorPct" :order="order" align="center">Errors</ThOrder>
          <ThOrder value="p50" :order="order" align="center">P50 latency</ThOrder>
          <ThOrder value="p99" :order="order" align="center">P99 latency</ThOrder>
        </tr>
      </thead>

      <thead v-show="loading">
        <tr class="v-data-table__progress">
          <th colspan="99" class="column">
            <v-progress-linear height="2" absolute indeterminate />
          </th>
        </tr>
      </thead>

      <tbody v-if="!items.length">
        <tr class="v-data-table__empty-wrapper">
          <td colspan="99">There are no data for the selected date range.</td>
        </tr>
      </tbody>

      <tbody>
        <tr v-for="(item, i) in items" :key="i">
          <td v-if="column" class="text-subtitle-1">
            <router-link :to="columnRoute(item[column])" @click.native.stop>
              {{ item[column] }}
            </router-link>
          </td>
          <td class="text-subtitle-1">
            <router-link :to="groupListRoute(item.system)" @click.native.stop>
              {{ item.system }}
            </router-link>
          </td>
          <td class="text-subtitle-2">
            <div class="d-flex align-center">
              <SparklineChart
                name="rate"
                :line="item.stats.rate"
                :time="item.stats.time"
                class="mr-2"
              />
              <XNum :value="item.rate" />
            </div>
          </td>
          <td class="text-subtitle-2">
            <div v-if="item.stats.errorCount" class="d-flex align-center">
              <SparklineChart
                name="errorPct"
                :line="item.stats.errorPct"
                :time="item.stats.time"
                class="mr-2"
              />
              {{ percent(item.errorPct) }}
            </div>
          </td>
          <td class="text-subtitle-2">
            <div v-if="item.stats.p50" class="d-flex align-center">
              <SparklineChart
                name="p50"
                :line="item.stats.p50"
                :time="item.stats.time"
                class="mr-2"
              />
              <XDuration :duration="item.p50" />
            </div>
          </td>
          <td class="text-subtitle-2">
            <div v-if="item.stats.p99" class="d-flex align-center">
              <SparklineChart
                name="p99"
                :line="item.stats.p99"
                :time="item.stats.time"
                class="mr-2"
              />
              <XDuration :duration="item.p99" />
            </div>
          </td>
        </tr>
      </tbody>
    </v-simple-table>
  </div>
</template>

<script lang="ts">
import { Route } from 'vue-router'
import { defineComponent, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { UseOrder } from '@/use/order'
import type { OverviewItem } from '@/use/system-stats'

// Components
import ThOrder from '@/components/ThOrder.vue'
import SparklineChart from '@/components/SparklineChart.vue'

// Utilities
import { xkey } from '@/models/otelattr'
import { quote } from '@/util/string'
import { percent } from '@/util/fmt'

export default defineComponent({
  name: 'OverviewTable',
  components: {
    ThOrder,
    SparklineChart,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    loading: {
      type: Boolean,
      required: true,
    },
    items: {
      type: Array as PropType<OverviewItem[]>,
      default: () => [],
    },
    order: {
      type: Object as PropType<UseOrder>,
      required: true,
    },
    column: {
      type: String,
      default: '',
    },
    attribute: {
      type: String,
      default: '',
    },
    baseColumnRoute: {
      type: Object as PropType<Route>,
      default: undefined,
    },
  },

  setup(props) {
    function columnRoute(value: string) {
      let where: string
      if (value === '') {
        where = `where ${props.column} not exists`
      } else {
        where = `where ${props.column} = ${quote(value)}`
      }

      const route = { ...props.baseColumnRoute }
      route.query = {
        ...route.query,
        query: `${route.query.query} | ${where}`,
      }

      return route
    }

    function groupListRoute(system: string) {
      return {
        name: 'SpanGroupList',
        query: {
          ...props.dateRange.queryParams(),
          system,
        },
      }
    }

    return {
      xkey,

      columnRoute,
      groupListRoute,
      percent,
    }
  },
})
</script>

<style lang="scss">
.v-data-table--large > .v-data-table__wrapper > table {
  & > tbody > tr > td {
    height: 60px;
  }
}
</style>
