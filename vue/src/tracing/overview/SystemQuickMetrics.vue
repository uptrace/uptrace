<template>
  <div>
    <v-row justify="center" class="metrics">
      <v-col v-if="showAll">
        <SystemQuickMetricCard :metric="metrics.all">
          <template #default="{ metric }">
            <XNum :value="metric.rate" :unit="Unit.Rate" title="{0} request per minute" />
          </template>
        </SystemQuickMetricCard>
      </v-col>

      <v-col v-if="metrics.http.rate">
        <SystemQuickMetricCard :metric="metrics.http">
          <template #default="{ metric }">
            <XNum :value="metric.rate" :unit="Unit.Rate" title="{0} request per minute" />
          </template>
        </SystemQuickMetricCard>
      </v-col>

      <v-col v-if="metrics.rpc.rate">
        <SystemQuickMetricCard :metric="metrics.rpc">
          <template #default="{ metric }">
            <XNum :value="metric.rate" :unit="Unit.Rate" title="{0} request per minute" />
          </template>
        </SystemQuickMetricCard>
      </v-col>

      <v-col v-if="metrics.db.rate">
        <SystemQuickMetricCard :metric="metrics.db">
          <template #default="{ metric }">
            <XNum :value="metric.rate" :unit="Unit.Rate" title="{0} query per minute" />
          </template>
        </SystemQuickMetricCard>
      </v-col>

      <v-col v-if="metrics.inMemDb.rate">
        <SystemQuickMetricCard :metric="metrics.inMemDb">
          <template #default="{ metric }">
            <XNum :value="metric.rate" :unit="Unit.Rate" title="{0} op per minute" />
          </template>
        </SystemQuickMetricCard>
      </v-col>

      <v-col v-if="metrics.failures.count">
        <SystemQuickMetricCard :metric="metrics.failures">
          <template #default="{ metric }">
            {{ ((metric.errorCount / metric.count) * 100).toFixed(2) }}%
          </template>
        </SystemQuickMetricCard>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import colors from 'vuetify/lib/util/colors'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { System } from '@/tracing/use-systems'

// Components
import SystemQuickMetricCard from '@/tracing/overview/SystemQuickMetricCard.vue'

// Utilities
import { AttrKey, isEventSystem } from '@/models/otelattr'
import { Unit } from '@/util/fmt'

export default defineComponent({
  name: 'SystemQuickMetrics',
  components: { SystemQuickMetricCard },

  props: {
    systems: {
      type: Array as PropType<System[]>,
      required: true,
    },
  },

  setup(props) {
    const metrics = computed(() => {
      const metrics = {
        all: {
          name: 'All',
          tooltip: 'Total number of spans per minute',
          color: colors.teal.base,
          rate: 0,
          suffix: 'span/min',
        },
        http: {
          name: 'HTTP',
          tooltip: 'Number of HTTP requests per minute',
          color: colors.blue.base,
          rate: 0,
          suffix: 'req/min',
        },
        rpc: {
          name: 'RPC',
          tooltip: 'Number of RPC requests per minute',
          color: colors.orange.base,
          rate: 0,
          suffix: 'req/min',
        },
        db: {
          name: 'Database',
          tooltip: 'Number of database queries per minute',
          color: colors.purple.base,
          rate: 0,
          suffix: 'op/min',
        },
        inMemDb: {
          name: 'In-memory DB',
          tooltip: 'Number of in-memory database commands per minute',
          color: colors.indigo.base,
          rate: 0,
          suffix: 'op/min',
        },
        failures: {
          name: 'Failures',
          tooltip: `Number of spans with ${AttrKey.spanStatusCode} = "error" divided by total number of spans`,
          color: colors.red.base,
          count: 0,
          errorCount: 0,
        },
      }

      for (let system of props.systems) {
        if (system.dummy) {
          continue
        }

        metrics.all.rate += system.rate

        if (!isEventSystem(system.system)) {
          metrics.failures.count += system.count
          metrics.failures.errorCount += system.errorCount
        }

        if (system.system.startsWith('http:')) {
          metrics.http.rate += system.rate
          continue
        }

        if (system.system.startsWith('rpc:')) {
          metrics.rpc.rate += system.rate
          continue
        }

        if (isInMemDb(system.system)) {
          metrics.inMemDb.rate += system.rate
          continue
        }

        if (isDb(system.system)) {
          metrics.db.rate += system.rate
          continue
        }
      }

      return metrics
    })

    const showAll = computed(() => {
      return metrics.value.http.rate === 0 && metrics.value.rpc.rate === 0
    })

    return { Unit, metrics, showAll }
  },
})

function isDb(system: string): boolean {
  return system.startsWith('db:')
}

function isInMemDb(system: string): boolean {
  switch (system) {
    case 'db:redis':
    case 'db:memcache':
      return true
  }
  return false
}
</script>

<style lang="scss" scoped>
.metrics ::v-deep .col {
  max-width: 250px;
}
</style>
