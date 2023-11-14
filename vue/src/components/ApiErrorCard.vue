<template>
  <v-container fluid>
    <v-row>
      <v-col class="text-center">
        <h1 v-if="error.statusCode" class="text-h1">{{ error.statusCode }}</h1>
        <h2 class="mt-4 text-h5">{{ error.message }}</h2>
      </v-col>
    </v-row>
    <v-row v-if="error.statusCode === 504" justify="center">
      <v-col cols="auto">
        <v-card max-width="700" flat>
          <p>
            Your query is taking a long time to complete. To improve the performance of your query,
            try the following:
          </p>

          <ul class="mb-4">
            <li>
              Narrow the date range, for example, select the "Last 1 hour" period if possible.
            </li>
            <li>
              Select a more specific system, for example, you can select the
              <code>http:all</code> system if you're analyzing HTTP requests.
            </li>
            <li>
              Further narrow the scope by adding the <code>.group_id</code> filter, for example,
              <code>where .group_id = 123456789</code>.
            </li>
            <li>
              Consider using
              <a href="https://uptrace.dev/get/enterprise.html">Uptrace Enterprise</a> edition which
              supports data pre-aggregation for common queries.
            </li>
          </ul>

          <p>
            You can also get in touch with us by
            <a href="mailto:support@uptrace.dev">sending an email</a> including the information
            below.
          </p>

          <div class="text--secondary">
            <div>Request id: {{ error.traceId }}</div>
            <div v-if="progress.rows">Rows read: <NumValue :value="progress.rows" /></div>
            <div v-if="progress.bytes">Bytes read: <BytesValue :value="progress.bytes" /></div>
            <div v-if="progress.totalRows">
              Total rows: <NumValue :value="progress.totalRows" />
            </div>
          </div>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { ApiError } from '@/use/promise'

interface Progress {
  rows: number
  bytes: number
  totalRows: number
}

export default defineComponent({
  name: 'ApiErrorCard',
  props: {
    error: {
      type: Object as PropType<ApiError>,
      required: true,
    },
  },

  setup(props) {
    const progress = computed((): Progress => {
      return props.error.data?.progress ?? {}
    })

    return { progress }
  },
})
</script>

<style lang="scss" scoped>
li {
  margin-bottom: 8px;
}
</style>
