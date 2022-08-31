<template>
  <div class="container--fixed-sm">
    <v-progress-linear v-if="loading" top absolute indeterminate></v-progress-linear>

    <PageToolbar>
      <v-toolbar-title>Send metrics to Uptrace</v-toolbar-title>

      <v-spacer />

      <ForceReloadBtn />
    </PageToolbar>

    <v-container fluid class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          <p>
            There are two types of
            <a href="https://uptrace.dev/opentelemetry/metrics.html" target="_blank">metrics</a>
            you can collect:
          </p>

          <ol class="mb-4">
            <li>
              <strong>In-app metrics</strong> using Uptrace client (OpenTelemetry distribution for
              Uptrace), for example, Go HTTP server metrics or user-defined metrics.
            </li>
            <li>
              <strong>Infrastructure metrics</strong> using OpenTelemetry Collector, for example,
              Linux/Windows system metrics or PostgreSQL metrics.
            </li>
          </ol>

          <p>
            Check our
            <a href="https://app.uptrace.dev/metrics/1">playground</a>
            to play with various types of metrics and
            <a href="https://uptrace.dev/opentelemetry/metrics.html" target="_blank">learn</a>
            how to create your own metrics.
          </p>
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar>
      <v-toolbar-title>In-app metrics</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          <p>
            To start collecting in-app metrics, you need to install Uptrace client and create
            instruments to report measurements. Use the following DSN to configure Uptrace client by
            following instructions for your programming language:
          </p>

          <PrismCode :code="`UPTRACE_DSN=${project.http.dsn}`" />
        </v-col>
      </v-row>

      <v-row>
        <v-col>
          <DistroIcons />
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar>
      <v-toolbar-title>Infrastructure metrics</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col class="text-subtitle-1">
          <p>
            To start monitoring your infrastructure, you need to install OpenTelemetry Collector on
            each host that you want to monitor. Collector acts as an agent that pulls metrics from
            monitored systems and exports them to Uptrace using the OTLP exporter.
          </p>

          <p>Use the following Uptrace project DSN to configure OpenTelemetry Collector:</p>

          <PrismCode :code="project.grpc.dsn" />
        </v-col>
      </v-row>

      <v-row>
        <v-col class="text-center">
          <v-btn
            color="primary"
            href="https://uptrace.dev/opentelemetry/collector.html#installation"
            target="_blank"
            >Install Collector</v-btn
          >
        </v-col>
      </v-row>
    </v-container>

    <PageToolbar :loading="loading">
      <v-toolbar-title>Supported software</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6 px-4 py-6">
      <v-row>
        <v-col>
          <SoftwareIcons />
        </v-col>
      </v-row>
    </v-container>

    <HelpLinks />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { useProject } from '@/use/project'

// Components
import ForceReloadBtn from '@/components/date/ForceReloadBtn.vue'
import DistroIcons from '@/components/DistroIcons.vue'
import SoftwareIcons from '@/components/SoftwareIcons.vue'
import HelpLinks from '@/components/HelpLinks.vue'

export default defineComponent({
  name: 'HelpCard',
  components: { ForceReloadBtn, DistroIcons, SoftwareIcons, HelpLinks },

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
  },

  setup() {
    const project = useProject()

    return { project }
  },
})
</script>

<style lang="scss" scoped></style>
