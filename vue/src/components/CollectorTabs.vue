<template>
  <v-card flat>
    <v-tabs v-model="tabs">
      <v-tab href="#grpc">GRPC</v-tab>
      <v-tab href="#http">HTTP</v-tab>
    </v-tabs>

    <v-tabs-items v-model="tabs">
      <v-tab-item value="grpc">
        <XCode language="yaml" :code="otlpGrpc"></XCode>
      </v-tab-item>
      <v-tab-item value="http">
        <XCode language="yaml" :code="otlpHttp"></XCode>
      </v-tab-item>
    </v-tabs-items>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, ref, computed } from '@vue/composition-api'

export default defineComponent({
  name: 'CollectorTabs',

  props: {
    otlp: {
      type: String,
      default: '',
    },
    dsn: {
      type: String,
      default: '',
    },
  },

  setup(props) {
    const tabs = ref()

    const otlpGrpc = computed(() => {
      return formatTemplate(otlpGrpcTpl.trim(), props.otlp, props.dsn)
    })

    const otlpHttp = computed(() => {
      return formatTemplate(otlpHttpTpl.trim(), props.otlp, props.dsn)
    })

    return { tabs, otlpGrpc, otlpHttp }
  },
})

function formatTemplate(format: string, ...args: any[]) {
  return format.replace(/{(\d+)}/g, function (match, number) {
    return typeof args[number] !== 'undefined' ? args[number] : match
  })
}

const otlpGrpcTpl = `
exporters:
  otlp:
    endpoint: {0}
    headers: {"uptrace-dsn": "{1}"}
    tls:
      insecure: true
`

const otlpHttpTpl = `
exporters:
  otlphttp:
    endpoint: {0}
    headers: {"uptrace-dsn": "{1}"}
    tls:
      insecure: true
`
</script>

<style lang="scss" scoped></style>
