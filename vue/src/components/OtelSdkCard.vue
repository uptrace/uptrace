<template>
  <div>
    <v-row>
      <v-col>
        <PrismCode :code="`export UPTRACE_DSN=&quot;${dsn}&quot;`" language="shell" />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <DistroIcons />
      </v-col>
    </v-row>

    <v-row>
      <v-col class="text-subtitle-1">
        If you are already using OpenTelemetry SDK with OTLP exporter, you can reconfigure it to
        export data to Uptrace using the following environment variables:
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <PrismCode :code="envVars" language="shell" />
      </v-col>
    </v-row>

    <v-row>
      <v-col>For more details, see the documentation for your programming language above.</v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Composables
import { useDsn } from '@/org/use-projects'

// Components
import DistroIcons from '@/components/DistroIcons.vue'

export default defineComponent({
  name: 'OtelSdkCard',
  components: { DistroIcons },

  props: {
    dsn: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const dsn = useDsn(computed(() => props.dsn))
    const envVars = computed(() => {
      return `
# Uncomment the appropriate protocol for your programming language.
# Only for OTLP/gRPC.
#export OTEL_EXPORTER_OTLP_ENDPOINT="${dsn.grpcEndpoint}"
# Only for OTLP/HTTP.
#export OTEL_EXPORTER_OTLP_ENDPOINT="${dsn.httpEndpoint}"

# Pass Uptrace DSN in gRPC/HTTP headers.
export OTEL_EXPORTER_OTLP_HEADERS="uptrace-dsn=${props.dsn}"

# Enable gzip compression.
export OTEL_EXPORTER_OTLP_COMPRESSION=gzip

# Enable exponential histograms.
export OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION=BASE2_EXPONENTIAL_BUCKET_HISTOGRAM

# Prefer delta temporality.
export OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=DELTA
      `.trim()
    })
    return { envVars }
  },
})
</script>

<style lang="scss" scoped></style>
