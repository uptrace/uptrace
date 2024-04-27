<template>
  <v-container fluid>
    <v-row>
      <v-col>
        <h2 class="mb-5 text-h5">Filtering</h2>

        <QueryExample query="where _kind = 'server'">
          <template #description>
            Filters start with <code>where</code> and look like SQL.
          </template>
        </QueryExample>

        <QueryExample query="where _name like '%foo%' | where user_name like '%foo%'">
          <template #description>
            Span fields start with an underscore. User-defined attribute names are unchanged.
          </template>
        </QueryExample>

        <QueryExample query="where display_name contains 'get|post'">
          <template #description>
            Search for spans that contain "get" or "post" (case-insensitive).
          </template>
        </QueryExample>

        <QueryExample query="where display_name like 'prefix%' or display_name like '%suffix'">
          <template #description>Filter spans that start or end with a prefix:</template>
        </QueryExample>

        <QueryExample query="where _status_code = 'error'">
          <template #description>
            Select failed spans using <code>.status_code</code> attribute.
          </template>
        </QueryExample>

        <QueryExample query="where log_message like '%group%does not exist%'">
          <template #description>
            Search logs using SQL LIKE operator (case-insensitive).
          </template>
        </QueryExample>

        <QueryExample query="where log_message ~ 'group(.*)does not exist'">
          <template #description> Search logs using a regexp (slower than like). </template>
        </QueryExample>

        <QueryExample query="where _duration > 10ms and _duration <= 0.5s">
          <template #description>Filter spans by duration.</template>
        </QueryExample>
      </v-col>
      <v-col>
        <h2 class="mb-5 text-h5">Grouping</h2>

        <QueryExample query="group by service_name, host_name">
          <template #description>Group by multiple columns.</template>
        </QueryExample>

        <QueryExample query="group by lower(service_name), upper(host_name)">
          <template #description>Lower/upper case the column value.</template>
        </QueryExample>

        <QueryExample query="group by lower(service_name) as service, upper(host_name) as host">
          <template #description>Aliases.</template>
        </QueryExample>

        <h2 class="mb-5 text-h5">Aggregation</h2>

        <QueryExample query="sum(_count) | per_min(sum(_count)) | _error_count | _error_rate">
          <template #description>Count number of spans and errors.</template>
        </QueryExample>

        <QueryExample query="sum(_count) | group by service_name, service_version">
          <template #description>Count number of spans for each service.</template>
        </QueryExample>

        <QueryExample query="group by service_name | p50(_duration)">
          <template #description>Select median span duration for each service.</template>
        </QueryExample>

        <QueryExample query="group by _group_id | top3(enduser_id)">
          <template #description>Select top 3 users for each group of spans.</template>
        </QueryExample>

        <QueryExample
          query="group by http_status_code | per_min(sum(_count)) | where http_status_code exists"
        >
          <template #description>
            Filter and group spans by <code>http_status_code</code> attribute.
          </template>
        </QueryExample>

        <QueryExample query="group by _group_id | uniq(enduser_id) | where _status_code = 'error'">
          <template #description>Count number of the affected users.</template>
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
