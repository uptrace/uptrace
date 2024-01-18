<template>
  <v-card max-width="800">
    <v-toolbar color="light-blue lighten-5" flat>
      <v-toolbar-title>{{ dashboard?.name || 'Dashboard' }}</v-toolbar-title>

      <v-spacer />

      <v-toolbar-items>
        <v-btn icon @click="$emit('click:cancel')">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-toolbar-items>
    </v-toolbar>

    <v-container>
      <PrismCode
        v-if="dash.yaml"
        :code="dash.yaml"
        language="yaml"
        target-style="max-height: 75vh; overflow-y: auto"
      />
      <v-skeleton-loader v-else type="article" />
    </v-container>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useYamlDashboard } from '@/metrics/use-dashboards'

// Misc
import { Dashboard } from '@/metrics/types'

export default defineComponent({
  name: 'DashboardYamlCard',

  props: {
    dashboard: {
      type: Object as PropType<Dashboard>,
      default: undefined,
    },
  },

  setup() {
    const dash = useYamlDashboard()
    return { dash }
  },
})
</script>

<style lang="scss" scoped></style>
