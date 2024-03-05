<template>
  <div>
    <v-row dense>
      <v-col>
        <v-alert type="info" prominent border="left" outlined class="mb-0">
          Don't forget to add the Uptrace exporter to <code>service.pipelines</code> section,
          because unused exporters are silently ignored.
        </v-alert>
      </v-col>
    </v-row>

    <v-row dense>
      <v-col>
        <v-tabs v-model="activeTab">
          <v-tab href="#grpc">GRPC</v-tab>
          <v-tab href="#http">HTTP</v-tab>
        </v-tabs>

        <v-tabs-items v-model="activeTab">
          <v-tab-item value="grpc">
            <PrismCode language="yaml" :code="otlpGrpc"></PrismCode>
          </v-tab-item>
          <v-tab-item value="http">
            <PrismCode language="yaml" :code="otlpHttp"></PrismCode>
          </v-tab-item>
        </v-tabs-items>
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        Note that the configuration above is minimal and only collects
        <a href="https://uptrace.dev/opentelemetry/collector-host-metrics.html" target="_blank"
          >host metrics</a
        >. To gather more metrics, you will need to configure additional receivers from the list
        below, for example,
        <a href="https://uptrace.dev/get/monitor/opentelemetry-postgresql.html" target="_blank"
          >PostgreSQL</a
        >
        or
        <a href="https://uptrace.dev/get/monitor/opentelemetry-mysql.html" target="_blank">MySQL</a>
        receivers.
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed } from 'vue'

// Composables
import { useDsn } from '@/org/use-projects'

export default defineComponent({
  name: 'CollectorTabs',

  props: {
    dsn: {
      type: String,
      default: '',
    },
  },

  setup(props) {
    const activeTab = shallowRef()
    const dsn = useDsn(computed(() => props.dsn))

    const otlpGrpc = computed(() => {
      return `
processors:
  resourcedetection:
    detectors: [env, system]
  cumulativetodelta:
  batch:
    send_batch_size: 10000
    timeout: 10s

receivers:
  otlp:
    protocols:
      grpc:
      http:
  hostmetrics:
    scrapers:
      cpu:
      disk:
      filesystem:
      load:
      memory:
      network:
      paging:

exporters:
  debug:
  otlp/uptrace:
    endpoint: ${dsn.grpcEndpoint}
    tls: { insecure: ${dsn.insecure} }
    headers:
      uptrace-dsn: '${props.dsn}'

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/uptrace, debug]
    metrics:
      receivers: [otlp]
      processors: [cumulativetodelta, batch]
      exporters: [otlp/uptrace, debug]
    metrics/host:
      receivers: [hostmetrics]
      processors: [cumulativetodelta, batch, resourcedetection]
      exporters: [otlp/uptrace, debug]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/uptrace, debug]
      `.trim()
    })

    const otlpHttp = computed(() => {
      return `
processors:
  resourcedetection:
    detectors: [env, system]
  cumulativetodelta:
  batch:
    send_batch_size: 10000
    timeout: 10s

receivers:
  otlp:
    protocols:
      grpc:
      http:
  hostmetrics:
    scrapers:
      cpu:
      disk:
      filesystem:
      load:
      memory:
      network:
      paging:

exporters:
  debug:
  otlphttp/uptrace:
    endpoint: ${dsn.httpEndpoint}
    headers:
      uptrace-dsn: '${props.dsn}'

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp/uptrace, debug]
    metrics:
      receivers: [otlp]
      processors: [cumulativetodelta, batch]
      exporters: [otlphttp/uptrace, debug]
    metrics/host:
      receivers: [hostmetrics]
      processors: [cumulativetodelta, batch, resourcedetection]
      exporters: [otlphttp/uptrace, debug]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp/uptrace, debug]
      `.trim()
    })

    return { activeTab, otlpGrpc, otlpHttp }
  },
})
</script>

<style lang="scss" scoped></style>
