<template>
  <div class="container--fixed-sm">
    <PageToolbar :loading="loading">
      <v-toolbar-title>Send data to Uptrace</v-toolbar-title>

      <v-spacer />

      <v-btn v-if="dateRange" small outlined @click="dateRange.reload">
        <v-icon small>mdi-refresh</v-icon>
        <span class="ml-1">Reload</span>
      </v-btn>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-body-1">
          <p>
            To start sending data to Uptrace, you need to install Uptrace client and configure it
            using the provided DSN (connection string).
          </p>

          <p>For Go and .NET use <strong>OTLP/gRPC</strong>:</p>

          <XCode :code="otlpGrpc" class="mb-4" />

          <p>For Python, Ruby, and Node.JS use <strong>OTLP/HTTP</strong>:</p>

          <XCode :code="otlpHttp" class="mb-4" />

          <p>
            For example, to run Go
            <a
              href="https://github.com/uptrace/uptrace-go/tree/master/example/basic"
              target="_blank"
              >basic</a
            >
            example:
          </p>

          <XCode :code="goCmd" language="go" class="mb-4" />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <DistroList />
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar :loading="loading">
      <v-toolbar-title>Already using OpenTelemetry Collector?</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-body-1">
          Uptrace natively supports OpenTelemetry Protocol (OTLP) in case you are already using
          OpenTelemetry Collector. Use the following OTLP exporter config:
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <CollectorTabs :http="otlpHttp" :grpc="otlpGrpc" />
        </v-col>
      </v-row>
    </v-container>

    <HelpLinks />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from '@vue/composition-api'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useWatchAxios } from '@/use/watch-axios'

// Components
import DistroList from '@/components/DistroList.vue'
import CollectorTabs from '@/components/CollectorTabs.vue'
import HelpLinks from '@/components/HelpLinks.vue'

export default defineComponent({
  name: 'HelpCard',
  components: { DistroList, CollectorTabs, HelpLinks },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    dateRange: {
      type: Object as PropType<UseDateRange>,
      default: undefined,
    },
  },

  setup() {
    const { data } = useWatchAxios(() => {
      return {
        url: '/api/tracing/conn-info',
      }
    })

    const otlpGrpc = computed(() => {
      return data.value?.grpc ?? 'http://localhost:4317'
    })

    const otlpHttp = computed(() => {
      return data.value?.http ?? 'http://localhost:14318'
    })

    const goCmd = computed(() => {
      return formatTemplate('UPTRACE_DSN={0} go run .', otlpGrpc.value)
    })

    return { otlpGrpc, otlpHttp, goCmd }
  },
})

function formatTemplate(format: string, ...args: any[]) {
  return format.replace(/{(\d+)}/g, function (match, number) {
    return typeof args[number] !== 'undefined' ? args[number] : match
  })
}
</script>

<style lang="scss" scoped></style>
