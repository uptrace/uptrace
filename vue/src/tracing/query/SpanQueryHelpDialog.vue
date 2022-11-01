<template>
  <v-dialog v-model="dialog" :max-width="1280" @keydown.esc="dialog = false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" v-bind="attrs" v-on="on">Help</v-btn>
    </template>

    <v-card>
      <v-toolbar flat color="blue lighten-5">
        <v-toolbar-title>Spans Querying Cheat Sheet</v-toolbar-title>
        <v-btn
          href="https://uptrace.dev/get/querying-spans.html"
          target="_blank"
          class="ml-6 primary"
        >
          <span>Documentation</span>
          <v-icon right>mdi-open-in-new</v-icon>
        </v-btn>
        <v-spacer />
        <v-toolbar-items>
          <v-btn icon @click="dialog = false"><v-icon>mdi-close</v-icon></v-btn>
        </v-toolbar-items>
      </v-toolbar>

      <v-container fluid class="pa-6">
        <v-row>
          <v-col>
            <h2 class="mb-5 text-h5">Filtering</h2>

            <SpanQueryExample query="where span.kind = 'server'">
              <template #description>
                Filters start with <code>where</code> and look like normal SQL.
              </template>
            </SpanQueryExample>

            <SpanQueryExample query="where span.name contains 'get|post'">
              <template #description>
                Search for spans that contain "get" or "post" (case-insensitive).
              </template>
            </SpanQueryExample>

            <SpanQueryExample query="where span.event_name contains 'timed out'">
              <template #description>
                Search for events that contain "timed out" (case-insensitive).
              </template>
            </SpanQueryExample>

            <SpanQueryExample query="where {span.name,span.event_name} contains 'timed out'">
              <template #description>
                You can also filter by multiple attributes at once.
              </template>
            </SpanQueryExample>

            <SpanQueryExample query="where status.code = 'error'">
              <template #description>
                Select failed spans using <code>status.code</code> attribute.
              </template>
            </SpanQueryExample>

            <SpanQueryExample query="where log.message like '%group%does not exist%'">
              <template #description>
                Search logs using SQL LIKE operator (case-insensitive).
              </template>
            </SpanQueryExample>

            <SpanQueryExample query="where span.duration > 10ms and span.duration <= 0.5s">
              <template #description>Filter spans by duration.</template>
            </SpanQueryExample>
          </v-col>
          <v-col>
            <h2 class="mb-5 text-h5">Aggregation</h2>

            <SpanQueryExample
              query="span.count | span.count_per_min | span.error_count | span.error_pct"
            >
              <template #description>Count number of spans and errors.</template>
            </SpanQueryExample>

            <SpanQueryExample query="span.count | group by service.name, service.version">
              <template #description>Count number of spans for each service.</template>
            </SpanQueryExample>

            <SpanQueryExample query="group by service.name | p50(span.duration)">
              <template #description>Select median span duration for each service.</template>
            </SpanQueryExample>

            <SpanQueryExample query="group by span.group_id | top3(enduser.id)">
              <template #description>Select top 3 users for each group of spans.</template>
            </SpanQueryExample>

            <SpanQueryExample
              query="group by http.status_code | span.count_per_min | where http.status_code exists"
            >
              <template #description>
                Filter and group spans by <code>http.status_code</code> attribute.
              </template>
            </SpanQueryExample>

            <SpanQueryExample
              query="group by span.group_id | uniq(enduser.id) | where span.status_code = 'error'"
            >
              <template #description>Count number of affected users.</template>
            </SpanQueryExample>
          </v-col>
        </v-row>

        <v-row>
          <v-spacer />
          <v-col cols="auto">
            <v-btn text color="primary" @click="dialog = false">Close</v-btn>
            <v-btn
              text
              color="primary"
              href="https://uptrace.dev/get/querying-spans.html"
              target="_blank"
            >
              <span>Read more</span>
              <v-icon right>mdi-open-in-new</v-icon>
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

// Components
import SpanQueryExample from '@/tracing/query/SpanQueryExample.vue'

export default defineComponent({
  name: 'SpanQueryHelpDialog',
  components: { SpanQueryExample },

  setup() {
    const dialog = shallowRef(false)
    return { dialog }
  },
})
</script>

<style lang="scss" scoped></style>
