<template>
  <div>
    <PageToolbar>
      <v-toolbar-title>See also</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6">
      <v-card flat>
        <v-card-text class="text-subtitle-1">
          <ul>
            <li>
              <a href="https://docs.uptrace.dev/" target="_blank">Uptrace documentation</a>
            </li>
            <li>
              <a href="https://opentelemetry.uptrace.dev/" target="_blank"
                >OpenTelemetry documentation</a
              >
            </li>
            <li>
              <a href="https://opentelemetry.uptrace.dev/instrumentations/" target="_blank"
                >Instrumentations</a
              >
            </li>
            <li>
              <a href="https://opentelemetry.uptrace.dev/collector/" target="_blank"
                >Collector Configurator</a
              >
            </li>
            <li>
              <a href="https://opentelemetry.uptrace.dev/attributes/" target="_blank"
                >Semantic attributes</a
              >
            </li>
            <li>
              <a href="https://github.com/uptrace" target="_blank">GitHub</a>
            </li>
          </ul>
        </v-card-text>

        <v-card-text class="text-body-1 text--primary">
          Have questions? Don't hesitate to <a href="mailto:support@uptrace.dev">email</a> us. We
          reply as soon as possible.
        </v-card-text>
      </v-card>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent } from '@vue/composition-api'

export default defineComponent({
  name: 'HelpLinks',
})
</script>
