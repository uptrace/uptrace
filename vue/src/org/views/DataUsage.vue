<template>
  <div class="container--fixed-sm">
    <portal to="navigation">
      <v-tabs :key="$route.fullPath" background-color="transparent">
        <v-tab :to="{ name: 'UserProfile' }" exact-path> Profile </v-tab>
        <v-tab :to="{ name: 'DataUsage' }" exact-path> Data usage </v-tab>
      </v-tabs>
    </portal>

    <PageToolbar>
      <v-toolbar-title v-if="usage.startTime && usage.endTime">
        <DateValue :value="usage.startTime" format="date" /> &mdash;
        <DateValue :value="usage.endTime" format="date" />
      </v-toolbar-title>
    </PageToolbar>

    <v-container class="py-6">
      <v-row class="text-center">
        <v-col cols="auto" class="px-6">
          <div class="text-h6 font-weight-regular">
            <BytesValue :value="usage.bytes" />
          </div>
          <div class="text-body-2 grey--text text--darken-1">bytes</div>
        </v-col>
        <v-col cols="auto" class="px-6">
          <div class="text-h6 font-weight-regular">
            <NumValue :value="usage.spans" />
          </div>
          <div class="text-body-2 grey--text text--darken-1">spans</div>
        </v-col>
        <v-col cols="auto" class="px-6">
          <div class="text-h6 font-weight-regular">
            <NumValue :value="usage.timeseries" />
          </div>
          <div class="text-body-2 grey--text text--darken-1">timeseries</div>
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <h3 class="text-subtitle-1 text--secondary">Ingested spans and logs</h3>
          <DataUsageChart
            :loading="usage.loading"
            :value="usage.data.spans"
            :time="usage.data.time"
            name="Spans & logs"
          />

          <h3 class="text-subtitle-1 text--secondary">Ingested bytes</h3>
          <DataUsageChart
            :loading="usage.loading"
            :value="usage.data.bytes"
            :time="usage.data.time"
            name="Bytes"
            unit="bytes"
          />

          <h3 class="text-subtitle-1 text--secondary">Timeseries</h3>
          <DataUsageChart
            :loading="usage.loading"
            :value="usage.data.timeseries"
            :time="usage.data.time"
            name="Timeseries"
          />
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composable
import { useTitle } from '@vueuse/core'
import { useDataUsage } from '@/org/use-data-usage'

// Components
import DataUsageChart from '@/org/DataUsageChart.vue'

export default defineComponent({
  name: 'DataUsage',
  components: { DataUsageChart },

  setup() {
    useTitle('Data Usage')

    const usage = useDataUsage()

    return {
      usage,
    }
  },
})
</script>

<style lang="scss" scoped></style>
