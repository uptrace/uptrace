<template>
  <div class="d-inline-block">
    <v-menu v-model="menu" offset-y>
      <template #activator="{ on, attrs }">
        <v-btn v-if="verbose" depressed small v-bind="attrs" v-on="on">
          <span>Monitor</span>
          <v-icon right>mdi-menu-down</v-icon>
        </v-btn>
        <v-btn v-else icon v-bind="attrs" v-on="on">
          <v-icon>mdi-dots-vertical</v-icon>
        </v-btn>
      </template>
      <v-list>
        <slot name="header-item" />

        <v-list-item v-for="(item, index) in menuItems" :key="index" :to="item.route">
          <v-list-item-title>{{ item.title }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { joinQuery } from '@/use/uql'
import { defaultMetricAlias } from '@/metrics/use-metrics'

// Misc
import { isEventSystem, isLogSystem } from '@/models/otel'

export default defineComponent({
  name: 'NewMonitorMenu',

  props: {
    systems: {
      type: Array as PropType<string[]>,
      required: true,
    },
    name: {
      type: String,
      required: true,
    },
    where: {
      type: String,
      default: undefined,
    },
    verbose: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const menu = shallowRef(false)

    const menuItems = computed(() => {
      if (isLogSystem(...props.systems)) {
        const metricName = 'uptrace_tracing_logs'
        return [
          {
            title: 'Monitor number of logs',
            route: routeFor(metricName, 'per_min($logs)'),
          },
        ]
      }

      if (isEventSystem(...props.systems)) {
        const metricName = 'uptrace_tracing_events'
        return [
          {
            title: 'Monitor number of events',
            route: routeFor(metricName, 'per_min($events)'),
          },
        ]
      }

      const metricName = 'uptrace_tracing_spans'
      return [
        {
          title: 'Monitor number of spans',
          route: routeFor(metricName, 'per_min(count($spans))'),
        },
        {
          title: 'Monitor number of failed spans',
          route: routeFor(metricName, 'per_min(count($spans{.status_code="error"}))'),
        },
        {
          title: 'Monitor error rate',
          route: routeFor(
            metricName,
            'count($spans{.status_code="error"}) / count($spans) as err_rate',
          ),
        },
        {
          title: 'Monitor p50 duration',
          route: routeFor(metricName, 'p50($spans)'),
        },
        {
          title: 'Monitor p90 duration',
          route: routeFor(metricName, 'p90($spans)'),
        },
        {
          title: 'Monitor p99 duration',
          route: routeFor(metricName, 'p99($spans)'),
        },
        {
          title: 'Monitor avg duration',
          route: routeFor(metricName, 'avg($spans)'),
        },
      ]
    })

    function routeFor(metricName: string, query: string) {
      return {
        name: 'MonitorMetricNew',
        query: {
          name: props.name,
          metric: metricName,
          alias: defaultMetricAlias(metricName),
          query: joinQuery([query, props.where]),
        },
      }
    }

    return {
      menu,
      menuItems,
    }
  },
})
</script>

<style lang="scss" scoped></style>
