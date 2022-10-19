<template>
  <div>
    <v-simple-table class="v-data-table--large">
      <thead class="v-data-table-header">
        <tr>
          <ThOrder v-if="attr" value="attr" :order="order">{{ attr }}</ThOrder>
          <ThOrder v-if="hasSystem" value="system" :order="order">System</ThOrder>
          <ThOrder value="rate" :order="order" align="center">Spans per minute</ThOrder>
          <ThOrder value="errorPct" :order="order" align="center">Error rate</ThOrder>
          <ThOrder value="durationP50" :order="order" align="center">P50 latency</ThOrder>
          <ThOrder value="durationP99" :order="order" align="center">P99 latency</ThOrder>
          <ThOrder value="durationMax" :order="order" align="end">Max</ThOrder>
          <th v-if="hasAction"></th>
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
          <td v-if="attr" class="text-subtitle-1">
            <router-link :to="itemRoute(item.attr)" @click.native.stop>
              <AnyValue :value="item.attr" :name="attr" />
            </router-link>
          </td>
          <td v-if="hasSystem" class="text-subtitle-1">
            <router-link :to="systemRoute(item.system)" @click.native.stop>
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
              <XNum :value="item.rate" :unit="Unit.Rate" title="{0} per minute" />
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
              <XPct :a="item.errorCount" :b="item.count" />
            </div>
          </td>
          <td class="text-subtitle-2">
            <div v-if="item.stats.durationP50" class="d-flex align-center">
              <SparklineChart
                name="p50"
                :line="item.stats.durationP50"
                :time="item.stats.time"
                class="mr-2"
              />
              <XDuration :duration="item.durationP50" />
            </div>
          </td>
          <td class="text-subtitle-2">
            <div v-if="item.stats.durationP99" class="d-flex align-center">
              <SparklineChart
                name="p99"
                :line="item.stats.durationP99"
                :time="item.stats.time"
                class="mr-2"
              />
              <XDuration :duration="item.durationP99" />
            </div>
          </td>
          <td class="text-subtitle-2 text-right">
            <XDuration v-if="item.durationMax !== undefined" :duration="item.durationMax" />
          </td>
          <td v-if="hasAction" class="text-center">
            <slot name="action" :item="item" />
          </td>
        </tr>
      </tbody>
    </v-simple-table>
  </div>
</template>

<script lang="ts">
import { Route } from 'vue-router'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useRoute } from '@/use/router'
import { UseOrder } from '@/use/order'
import type { OverviewItem } from '@/tracing/overview/types'

// Components
import ThOrder from '@/components/ThOrder.vue'
import SparklineChart from '@/components/SparklineChart.vue'

// Utilities
import { isEventSystem } from '@/models/otelattr'
import { Unit } from '@/util/fmt'
import { quote } from '@/util/string'

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
    attr: {
      type: String,
      default: '',
    },
    baseItemRoute: {
      type: Object as PropType<Route>,
      default: undefined,
    },
  },

  setup(props, ctx) {
    const route = useRoute()

    const hasAction = computed(() => {
      return 'action' in ctx.slots
    })

    const hasSystem = computed(() => {
      return props.items.some((item) => item.system)
    })

    function itemRoute(value: string) {
      let where: string
      if (value === '') {
        where = `where ${props.attr} not exists`
      } else {
        where = `where ${props.attr} = ${quote(value)}`
      }

      const route = { ...props.baseItemRoute }
      route.query = {
        ...route.query,
        query: `${route.query.query} | ${where}`,
      }

      return route
    }

    function systemRoute(system: string) {
      return {
        name: isEventSystem(system) ? 'LogGroupList' : 'SpanGroupList',
        query: {
          ...route.value.query,
          system,
        },
      }
    }

    return {
      Unit,
      hasAction,
      hasSystem,

      itemRoute,
      systemRoute,
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
