<template>
  <v-dialog v-model="dialog" max-width="1200">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" v-bind="attrs" v-on="on">Help</v-btn>
    </template>

    <v-card>
      <v-toolbar flat color="blue lighten-5">
        <v-toolbar-title>Uptrace Metrics Cheat Sheet</v-toolbar-title>
        <v-spacer />
        <v-btn icon @click="dialog = false"><v-icon>mdi-close</v-icon></v-btn>
      </v-toolbar>

      <v-container fluid class="pa-6">
        <v-row>
          <v-col>
            <h2 class="mb-5 text-h5">Filtering timeseries</h2>

            <MetricQueryExample query="$metric1 | $metric2">
              <template #description
                >Metric names start with <code>$</code>. Each expr is separated with
                <code>|</code>.</template
              >
            </MetricQueryExample>

            <MetricQueryExample query='$cpu_time{cpu="0",mode="idle"}'>
              <template #description>Select timeseries with given attributes.</template>
            </MetricQueryExample>

            <MetricQueryExample query='$cpu_time{cpu="0",mode="idle"} as first_cpu_idle_time'>
              <template #description>Give timeseries a shorter or better name.</template>
            </MetricQueryExample>

            <MetricQueryExample query='$cpu_time{cpu!="0",mode~"user|system"}'>
              <template #description
                >Equal <code>=</code>, not equal <code>!=</code>, regexp match <code>~</code>,
                regexp no match <code>!~</code>.</template
              >
            </MetricQueryExample>

            <MetricQueryExample query="$cache_hits | $cache_misses | where host.name = localhost">
              <template #description
                >Filter all timeseries at once by <code>host.name</code>.</template
              >
            </MetricQueryExample>
          </v-col>

          <v-col>
            <h2 class="mb-5 text-h5">Grouping and combining</h2>

            <MetricQueryExample query="$hits | $misses | group by host.name">
              <template #description>Select cache hits and misses on every hostname.</template>
            </MetricQueryExample>

            <MetricQueryExample query="$hits + $misses | group by host.name">
              <template #description>Sum timeseries with matching attributes.</template>
            </MetricQueryExample>

            <MetricQueryExample
              query="$cache{type=hits} as hits | $cache{type=misses} as misses | hits + misses as sum"
            >
              <template #description>Combine timeseries using the aliases.</template>
            </MetricQueryExample>

            <MetricQueryExample
              query="$metric1 group by service.name | $metric2 group by host.name"
            >
              <template #description>Individual grouping for each timeseries.</template>
            </MetricQueryExample>

            <MetricQueryExample query="$metric_name group by all">
              <template #description>Group by all attributes like Prometheus.</template>
            </MetricQueryExample>
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <h2 class="mb-5 text-h5">Counter Instrument: $cache</h2>

            <MetricQueryExample query="$cache{type=hits}">
              <template #description>Select number of cache hits.</template>
            </MetricQueryExample>

            <MetricQueryExample query="$cache{type=misses}">
              <template #description>Number of cache misses.</template>
            </MetricQueryExample>

            <MetricQueryExample query="$cache{type=hits} + $cache{type=misses}">
              <template #description>Sum of cache hits and misses.</template>
            </MetricQueryExample>

            <MetricQueryExample query="$cache">
              <template #description
                >Sum of cache hits, misses, and possibly other types/timeseries.</template
              >
            </MetricQueryExample>

            <MetricQueryExample query="per_min($cache) | per_sec($cache)">
              <template #description>
                Number of cache operations per minute and per second.
              </template>
            </MetricQueryExample>

            <MetricQueryExample query="per_min($cache) group by type">
              <template #description
                >Number of cache operations per minute grouped by type.</template
              >
            </MetricQueryExample>
          </v-col>

          <v-col>
            <h2 class="mb-5 text-h5">Histogram Instrument: $srv_duration</h2>

            <MetricQueryExample query="p50($srv_duration)">
              <template #description>P50 duration.</template>
            </MetricQueryExample>

            <MetricQueryExample query="p90($srv_duration{env=prod}) | p90($srv_duration{env=dev})">
              <template #description
                >P90 duration in <code>prod</code> and <code>env</code> environments.</template
              >
            </MetricQueryExample>

            <MetricQueryExample query="avg($srv_duration) group by host.name">
              <template #description>Avg duration on each hostname.</template>
            </MetricQueryExample>

            <MetricQueryExample query='avg($srv_duration{host.name~"api\d+$"})'>
              <template #description>Avg duration on hostnames matching the regexp.</template>
            </MetricQueryExample>

            <MetricQueryExample query="per_min($srv_duration) | per_sec($srv_duration)">
              <template #description>Number of requests per minute and per second.</template>
            </MetricQueryExample>

            <MetricQueryExample query="min($srv_duration) | max($srv_duration)">
              <template #description>Min and max duration.</template>
            </MetricQueryExample>
          </v-col>
        </v-row>

        <v-row>
          <v-spacer />
          <v-col cols="auto">
            <v-btn text color="primary" @click="dialog = false">Close</v-btn>
            <v-btn
              text
              color="primary"
              href="https://uptrace.dev/docs/querying-metrics.html"
              target="_blank"
              >Read more</v-btn
            >
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

import MetricQueryExample from '@/metrics/query/MetricQueryExample.vue'

export default defineComponent({
  name: 'MetricQueryHelpDialog',
  components: { MetricQueryExample },

  setup() {
    const dialog = shallowRef(false)
    return { dialog }
  },
})
</script>

<style lang="scss" scoped></style>
