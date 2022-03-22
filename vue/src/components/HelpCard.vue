<template>
  <div class="container--fixed-sm">
    <PageToolbar :loading="loading">
      <v-toolbar-title>Send data to Uptrace</v-toolbar-title>

      <v-spacer />

      <DateRangePicker v-if="dateRange" :date-range="dateRange" />
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-body-1">
          <p>
            To start sending data to Uptrace, you need to install Uptrace client and configure it
            using the provided DSN (connection string).
          </p>

          <p>For Go, Python, and .NET use <strong>OTLP/gRPC</strong>:</p>

          <XCode :code="grpcDsn" class="mb-4" />

          <p>For Ruby and Node.JS use <strong>OTLP/HTTP</strong>:</p>

          <XCode :code="httpDsn" class="mb-4" />

          <p>
            See
            <a href="https://get.uptrace.dev/guide/#otlp" target="_blank">documentation</a>
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
import DateRangePicker from '@/components/DateRangePicker.vue'
import CollectorTabs from '@/components/CollectorTabs.vue'
import HelpLinks from '@/components/HelpLinks.vue'

export default defineComponent({
  name: 'HelpCard',
  components: { DateRangePicker, CollectorTabs, HelpLinks },

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

    return { grpcEndpoint, httpEndpoint, grpcDsn, httpDsn }
  },
})
</script>

<style lang="scss" scoped></style>
