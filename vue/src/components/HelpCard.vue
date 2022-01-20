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
            See
            <a href="https://docs.uptrace.dev/guide/os.html#otlp" target="_blank">documentation</a>
            for instructions for your programming language.
          </p>
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
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useWatchAxios } from '@/use/watch-axios'

// Components
import CollectorTabs from '@/components/CollectorTabs.vue'
import HelpLinks from '@/components/HelpLinks.vue'

export default defineComponent({
  name: 'HelpCard',
  components: { CollectorTabs, HelpLinks },

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

    const otlpGrpc = computed(() => {
      return data.value?.grpc ?? 'http://localhost:4317'
    })

    const otlpHttp = computed(() => {
      return data.value?.http ?? 'http://localhost:14318'
    })

    return { otlpGrpc, otlpHttp }
  },
})
</script>

<style lang="scss" scoped></style>
