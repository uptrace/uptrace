<template>
  <v-container fluid>
    <v-row>
      <v-col>
        <h2 class="mb-5 text-h5">Filtering</h2>

        <QueryExample query="where .kind = 'server'">
          <template #description>
            Filters start with <code>where</code> and look like SQL.
          </template>
        </QueryExample>

        <QueryExample query="where .name like '%foo%' | where user.name like '%foo%'">
          <template #description>
            Span fields start with a dot. User-defined attribute names are unchanged.
          </template>
        </QueryExample>

        <QueryExample query="where display.name contains 'get|post'">
          <template #description>
            Search for spans that contain "get" or "post" (case-insensitive).
          </template>
        </QueryExample>

        <QueryExample query="where display.name like 'prefix%' or display.name like '%suffix'">
          <template #description>Filter spans that start or end with a prefix:</template>
        </QueryExample>

        <QueryExample query="where {.name,.event_name} contains 'timed out'">
          <template #description> You can also filter by multiple attributes at once. </template>
        </QueryExample>

        <QueryExample query="where .status_code = 'error'">
          <template #description>
            Select failed spans using <code>.status_code</code> attribute.
          </template>
        </QueryExample>

        <QueryExample query="where log.message like '%group%does not exist%'">
          <template #description>
            Search logs using SQL LIKE operator (case-insensitive).
          </template>
        </QueryExample>

        <QueryExample query="where log.message ~ 'group(.*)does not exist'">
          <template #description> Search logs using a regexp (slower than like). </template>
        </QueryExample>

        <QueryExample query="where .duration > 10ms and .duration <= 0.5s">
          <template #description>Filter spans by duration.</template>
        </QueryExample>
      </v-col>
      <v-col>
        <h2 class="mb-5 text-h5">Aggregation</h2>

        <QueryExample query=".count | per_min(.count) | .error_count | .error_rate">
          <template #description>Count number of spans and errors.</template>
        </QueryExample>

        <QueryExample query=".count | group by service.name, service.version">
          <template #description>Count number of spans for each service.</template>
        </QueryExample>

        <QueryExample query="group by service.name | p50(.duration)">
          <template #description>Select median span duration for each service.</template>
        </QueryExample>

        <QueryExample query="group by .group_id | top3(enduser.id)">
          <template #description>Select top 3 users for each group of spans.</template>
        </QueryExample>

        <QueryExample
          query="group by http.status_code | per_min(.count) | where http.status_code exists"
        >
          <template #description>
            Filter and group spans by <code>http.status_code</code> attribute.
          </template>
        </QueryExample>

        <QueryExample query="group by .group_id | uniq(enduser.id) | where .status_code = 'error'">
          <template #description>Count number of affected users.</template>
        </QueryExample>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Components
import QueryExample from '@/components/QueryExample.vue'

export default defineComponent({
  name: 'CheatSheetCard',
  components: { QueryExample },
})
</script>

<style lang="scss" scoped></style>
