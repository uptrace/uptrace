<template>
  <div>
    <PageToolbar>
      <v-toolbar-title>See also</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6">
      <v-card flat>
        <v-card-text class="text-subtitle-1 text--primary">
          <ul>
            <li>
              <a href="https://uptrace.dev/docs/" target="_blank">Documentation</a> - getting
              started with OpenTelemetry distro for Uptrace, for example, uptrace-go or
              uptrace-python.
            </li>
            <li>
              <a href="https://uptrace.dev/opentelemetry/" target="_blank">OpenTelemetry</a> -
              OpenTelemetry knowledge base, for example,
              <a href="https://uptrace.dev/opentelemetry/distributed-tracing.html" target="_blank"
                >tracing</a
              >
              and
              <a href="https://uptrace.dev/opentelemetry/metrics.html" target="_blank">metrics</a>.
            </li>
            <li>
              <a href="https://uptrace.dev/opentelemetry/instrumentations/" target="_blank"
                >Instrumentations</a
              >
              - a list of OpenTelemetry instrumentations.
            </li>
            <li>
              <a href="https://uptrace.dev/opentelemetry/collector-config.html" target="_blank"
                >Collector Configurator</a
              >
              - a single page application to help you configure OpenTelemetry Collector.
            </li>
            <li>
              <a href="https://uptrace.dev/opentelemetry/attributes.html" target="_blank"
                >Semantic Attributes</a
              >
              - a conventional list of semantic attributes recommended by OpenTelemetry.
            </li>
            <li>
              <a href="https://github.com/uptrace" target="_blank">Uptrace on GitHub</a>
            </li>
          </ul>
        </v-card-text>

        <v-card-text class="text-subtitle-1 text--primary">
          Have questions? Don't hesitate to <a href="mailto:support@uptrace.dev">email</a> us. We
          reply as soon as possible.
        </v-card-text>
      </v-card>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'HelpLinks',
})
</script>

<style lang="scss" scoped>
li {
  margin-bottom: 5px;
}
</style>
