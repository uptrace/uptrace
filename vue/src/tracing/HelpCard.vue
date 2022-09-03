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

          <p>For Go, Python, and .NET, use <strong>OTLP/gRPC</strong> (port 14317):</p>

          <PrismCode :code="`UPTRACE_DSN=${project.grpc.dsn}`" class="mb-4" />

          <p>For Ruby and Node.JS, use <strong>OTLP/HTTP</strong> (port 14318):</p>

          <PrismCode :code="`UPTRACE_DSN=${project.http.dsn}`" class="mb-4" />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <DistroIcons />
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
          <a href="https://uptrace.dev/opentelemetry/collector.html#configuration" target="_blank"
            >config</a
          >:
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <CollectorTabs :http="project.http" :grpc="project.grpc" />
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar :loading="loading">
      <v-toolbar-title>Vector Logs</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          To configure Vector to send logs to Uptrace, use the HTTP sink and pass your project DSN
          via HTTP headers. For example, to collect syslog messages you can create the following
          Vector config:
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <PrismCode language="toml" :code="vectorConfig" />
        </v-col>
      </v-row>

      <v-row>
        <v-col class="text-subtitle-1">
          See
          <a href="https://uptrace.dev/opentelemetry/structured-logging.html" target="_blank"
            >documentation</a
          >
          for more details.
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar :loading="loading">
      <v-toolbar-title>Zipkin API</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          Uptrace also supports Zipkin JSON API at <code>/api/v2/spans</code>:
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <PrismCode language="bash" :code="zipkinCurl" />
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useProject } from '@/use/project'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import CollectorTabs from '@/components/CollectorTabs.vue'
import DistroIcons from '@/components/DistroIcons.vue'
import FrameworkIcons from '@/components/FrameworkIcons.vue'

export default defineComponent({
  name: 'HelpCard',
  components: {
    DateRangePicker,
    CollectorTabs,
    DistroIcons,
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
    const project = useProject()

    const vectorConfig = computed(() => {
      return `
[sources.in]
type = "file"
include = ["/var/log/syslog"]

[sinks.out]
type = "http"
inputs = ["in"]
encoding.codec = "ndjson"
compression = "gzip"
uri = "${project.http.endpoint}/api/v1/vector/logs"
headers.uptrace-dsn = "${project.http.dsn}"
      `.trim()
    })

    const zipkinCurl = computed(() => {
      return `
curl -X POST '${project.http.endpoint}/api/v2/spans' \\
  -H 'Content-Type: application/json' \\
  -H 'uptrace-dsn: ${project.http.dsn}' \\
  -d @spans.json
      `.trim()
    })

    return { project, vectorConfig, zipkinCurl }
  },
})
</script>

<style lang="scss" scoped></style>
