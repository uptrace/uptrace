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
              <a href="https://uptrace.dev/get/" target="_blank">Documentation</a> to get you
              started with OpenTelemetry distros for Uptrace.
            </li>
            <li>
              <a href="https://uptrace.dev/opentelemetry/" target="_blank"
                >OpenTelemetry knowledge base</a
              >, for example,
              <a href="https://uptrace.dev/opentelemetry/distributed-tracing.html" target="_blank"
                >tracing</a
              >
              and
              <a href="https://uptrace.dev/opentelemetry/metrics.html" target="_blank">metrics</a>.
            </li>
            <li>
              <a href="https://uptrace.dev/opentelemetry/instrumentations/" target="_blank"
                >Instrumentations</a
              >. A list of OpenTelemetry instrumentations.
            </li>
            <li>
              <a href="https://uptrace.dev/opentelemetry/attributes.html" target="_blank"
                >Semantic Attributes</a
              >. A conventional list of semantic attributes recommended by OpenTelemetry.
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
