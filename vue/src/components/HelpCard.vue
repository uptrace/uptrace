<template>
  <div class="container--fixed-sm">
    <PageToolbar :loading="loading">
      <v-toolbar-title>Send data to Uptrace</v-toolbar-title>

      <v-spacer />

      <DateRangePicker v-if="dateRange" :date-range="dateRange" />
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          <p>
            To start sending data to Uptrace, you need to install OpenTelemetry distro for Uptrace
            and configure it using the provided DSN (connection string).
          </p>

          <p>For Go, Python, and .NET use <strong>OTLP/gRPC</strong> (port 14317):</p>

          <XCode :code="`UPTRACE_DSN=${grpcDsn}`" class="mb-4" />

          <p>For Ruby and Node.JS use <strong>OTLP/HTTP</strong> (port 14318):</p>

          <XCode :code="`UPTRACE_DSN=${httpDsn}`" class="mb-4" />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <UptraceDistroIcons />
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar :loading="loading">
      <v-toolbar-title>Quickstart</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <FrameworkIcons />
    </v-container>

    <PageToolbar :loading="loading">
      <v-toolbar-title>Already using OpenTelemetry Collector?</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          Uptrace natively supports OpenTelemetry Protocol (OTLP) in case you are already using
          OpenTelemetry Collector. Use the following OTLP exporter
          <a
            href="https://opentelemetry.uptrace.dev/guide/collector.html#configuration"
            target="_blank"
            >config</a
          >:
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <CollectorTabs
            :http-endpoint="httpEndpoint"
            :grpc-endpoint="grpcEndpoint"
            :http-dsn="httpDsn"
            :grpc-dsn="grpcDsn"
          />
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar :loading="loading">
      <v-toolbar-title>Already using Zipkin API?</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          Uptrace also supports Zipkin JSON API at <code>/api/v2/spans</code>:
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <XCode language="bash" :code="zipkinCurl" />
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useWatchAxios } from '@/use/watch-axios'

// Components
import DateRangePicker from '@/components/DateRangePicker.vue'
import CollectorTabs from '@/components/CollectorTabs.vue'
import UptraceDistroIcons from '@/components/UptraceDistroIcons.vue'
import FrameworkIcons from '@/components/FrameworkIcons.vue'

export default defineComponent({
  name: 'HelpCard',
  components: {
    DateRangePicker,
    CollectorTabs,
    UptraceDistroIcons,
    FrameworkIcons,
  },

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
    const { route } = useRouter()

    const { data } = useWatchAxios(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/tracing/${projectId}/conn-info`,
      }
    })

    const grpcEndpoint = computed(() => {
      return data.value?.grpc?.endpoint ?? 'http://localhost:14317'
    })

    const httpEndpoint = computed(() => {
      return data.value?.http?.endpoint ?? 'http://localhost:14318'
    })

    const grpcDsn = computed(() => {
      return data.value?.grpc?.dsn ?? 'http://localhost:14317'
    })

    const httpDsn = computed(() => {
      return data.value?.http?.dsn ?? 'http://localhost:14318'
    })

    return { grpcEndpoint, httpEndpoint, grpcDsn, httpDsn, zipkinCurl }
  },
})

const zipkinCurl = `
curl -X POST 'http://localhost:14318/api/v2/spans' -H 'Content-Type: application/json' -d @spans.json
`.trim()
</script>

<style lang="scss" scoped></style>
