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
import { useRoute } from '@/use/router'
import { useYamlDashboard } from '@/metrics/use-dashboards'
import { injectForceReload } from '@/use/force-reload'

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

  setup(props) {
    const route = useRoute()
    const forceReload = injectForceReload()

    const dash = useYamlDashboard(() => {
      if (!props.dashboard.id) {
        return
      }

      const { projectId } = route.value.params
      return {
        url: `/internal/v1/metrics/${projectId}/dashboards/${props.dashboard.id}/yaml`,
        params: forceReload.params,
      }
    })

    return { dash }
  },
})
</script>

<style lang="scss" scoped></style>
