<template>
  <div>
    <v-row v-if="loading">
      <v-col v-for="i in 8" :key="i">
        <v-skeleton-loader type="card" height="100px"></v-skeleton-loader>
      </v-col>
    </v-row>

    <v-row v-else justify="center" :dense="$vuetify.breakpoint.mdAndDown" class="metrics">
      <template v-for="(metric, metricName) in metrics">
        <v-col v-if="metric.count || metric.rate" :key="metricName" cols="auto">
          <SystemQuickMetricCard :metric="metric">
            <template v-if="metricName === 'failures'" #default="{ metric }">
              <XPct :a="metric.errorCount" :b="metric.count" />
            </template>
          </SystemQuickMetricCard>
        </v-col>
      </template>
    </v-row>
  </div>
</template>

<script lang="ts">
import colors from 'vuetify/lib/util/colors'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { System } from '@/tracing/system/use-systems'

// Components
import SystemQuickMetricCard from '@/tracing/system/SystemQuickMetricCard.vue'

// Utilities
import { AttrKey, isEventSystem } from '@/models/otel'
import { Unit } from '@/util/fmt'

export default defineComponent({
  name: 'SystemQuickMetrics',
  components: { SystemQuickMetricCard },

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    systems: {
      type: Array as PropType<System[]>,
      required: true,
    },
  },

  setup(props) {
    const metrics = computed(() => {
      const metrics = {
        all: {
          name: 'Spans',
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
        logs: {
          name: 'Logs',
          tooltip: 'Number of logs per minute',
          rate: 0,
          suffix: 'log/min',
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

        if (system.system.startsWith('log:')) {
          metrics.logs.rate += system.rate
          continue
        }
      }

      return metrics
    })

    return { Unit, metrics }
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

<style lang="scss" scoped></style>
