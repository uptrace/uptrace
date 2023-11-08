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

exporters:
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
      exporters: [otlp/uptrace]
    metrics:
      receivers: [otlp]
      processors: [cumulativetodelta, batch, resourcedetection]
      exporters: [otlp/uptrace]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/uptrace]
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

exporters:
  otlphttp/uptrace:
    endpoint: ${dsn.httpEndpoint}
    headers:
      uptrace-dsn: '${props.dsn}'

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp/uptrace]
    metrics:
      receivers: [otlp]
      processors: [cumulativetodelta, batch, resourcedetection]
      exporters: [otlphttp/uptrace]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp/uptrace]
      `.trim()
    })

    return { activeTab, otlpGrpc, otlpHttp }
  },
})
</script>

<style lang="scss" scoped></style>
