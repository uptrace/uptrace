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

        <v-list-item v-for="(item, index) in items" :key="index" :to="item.route">
          <v-list-item-title>{{ item.title }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed } from 'vue'

export default defineComponent({
  name: 'NewMonitorMenu',

  props: {
    metric: {
      type: String,
      default: 'uptrace.tracing.spans',
    },
    name: {
      type: String,
      required: true,
    },
    axiosParams: {
      type: Object,
      default: undefined,
    },
    where: {
      type: String,
      required: true,
    },
    eventsMode: {
      type: Boolean,
      default: false,
    },
    verbose: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const menu = shallowRef(false)

    const items = computed(() => {
      if (props.eventsMode) {
        return [
          {
            title: 'Monitor number of events',
            route: eventsRoute('per_min($events)'),
          },
        ]
      }

      return [
        {
          title: 'Monitor number of spans',
          route: spansRoute('per_min($spans)'),
        },
        {
          title: 'Monitor number of failed spans',
          route: spansRoute('per_min($spans{span.status_code="error"})'),
        },
        {
          title: 'Monitor error rate',
          route: spansRoute('count($spans{span.status_code="error"}) / count($spans) as err_rate'),
        },
        {
          title: 'Monitor p50 duration',
          route: spansRoute('p50($spans)'),
        },
        {
          title: 'Monitor p90 duration',
          route: spansRoute('p90($spans)'),
        },
        {
          title: 'Monitor p99 duration',
          route: spansRoute('p99($spans)'),
        },
        {
          title: 'Monitor avg duration',
          route: spansRoute('avg($spans)'),
        },
      ]
    })

    function spansRoute(query: string) {
      return {
        name: 'MonitorMetricNew',
        query: {
          name: props.name,
          metric: props.metric,
          alias: 'spans',
          query: `${query} | ${props.where}`,
        },
      }
    }

    function eventsRoute(query: string) {
      return {
        name: 'MonitorMetricNew',
        query: {
          name: props.name,
          metric: 'uptrace.tracing.events',
          alias: 'events',
          query: `${query} | ${props.where}`,
        },
      }
    }

    return {
      menu,
      items,
    }
  },
})
</script>

<style lang="scss" scoped></style>
