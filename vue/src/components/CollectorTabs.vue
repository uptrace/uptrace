<template>
  <v-card flat>
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
  </v-card>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { ConnDetails } from '@/use/project'

export default defineComponent({
  name: 'CollectorTabs',

  props: {
    grpc: {
      type: Object as PropType<ConnDetails>,
      required: true,
    },
    http: {
      type: Object as PropType<ConnDetails>,
      required: true,
    },
  },

  setup(props) {
    const activeTab = shallowRef('')

    const otlpGrpc = computed(() => {
      return `
exporters:
  otlp:
    endpoint: ${props.grpc.endpoint}
    tls:
      insecure: true
    headers:
      uptrace-dsn: '${props.grpc.dsn}'
      `.trim()
    })

    const otlpHttp = computed(() => {
      return `
exporters:
  otlphttp:
    endpoint: ${props.http.endpoint}
    tls:
      insecure: true
    headers:
      uptrace-dsn: '${props.http.dsn}'
      `.trim()
    })

    return { activeTab, otlpGrpc, otlpHttp }
  },
})
</script>

<style lang="scss" scoped></style>
