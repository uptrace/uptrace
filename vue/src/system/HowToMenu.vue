<template>
  <v-menu open-on-hover offset-x right :nudge-left="1">
    <template #activator="{ attrs, on }">
      <v-list-item ripple v-bind="attrs" v-on="on">
        <v-list-item-action>
          <v-icon>mdi-help-circle-outline</v-icon>
        </v-list-item-action>
        <v-list-item-content>
          <v-list-item-title>How to?</v-list-item-title>
        </v-list-item-content>
        <v-list-item-icon>
          <v-icon dense>mdi-chevron-right</v-icon>
        </v-list-item-icon>
      </v-list-item>
    </template>

    <v-sheet max-width="550">
      <v-list>
        <v-list-item :to="{ name: 'ProjectDsn', params }">
          <v-list-item-content>
            <v-list-item-title>How to find my project DSN?</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item :to="{ name: 'TracingHelp', params }">
          <v-list-item-content>
            <v-list-item-title>How to setup tracing?</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item :to="{ name: 'MetricsHelp', params }">
          <v-list-item-content>
            <v-list-item-title>How to collect metrics?</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item href="https://uptrace.dev/get/logging.html" target="_blank">
          <v-list-item-content>
            <v-list-item-title>How to monitor logs?</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item :to="{ name: 'TracingCheatsheet', params }">
          <v-list-item-content>
            <v-list-item-title>How to query spans?</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item :to="{ name: 'MetricsCheatsheet', params }">
          <v-list-item-content>
            <v-list-item-title>How to query metrics?</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-sheet>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { Project } from '@/org/use-projects'

export default defineComponent({
  name: 'ProjectMenu',

  props: {
    project: {
      type: Object as PropType<Project>,
      required: true,
    },
  },

  setup(props) {
    const params = computed(() => {
      return { projectId: props.project.id }
    })
    return { params }
  },
})
</script>

<style lang="scss" scoped></style>
