<template>
  <span>
    <v-btn-toggle v-model="view" mandatory dense borderless>
      <v-btn
        value="grid"
        :loading="dashMan.pending"
        :disabled="dashboard.isTemplate"
        small
        title="Grid view"
      >
        <span class="mr-1 hidden-sm-and-down">Grid</span>
        <v-icon small>mdi-view-grid-outline</v-icon>
      </v-btn>
      <v-btn
        value="table"
        :loading="dashMan.pending"
        :disabled="dashboard.isTemplate"
        small
        title="Table view"
      >
        <span class="mr-1 hidden-sm-and-down">Table</span>
        <v-icon small>mdi-view-sequential-outline</v-icon>
      </v-btn>
    </v-btn-toggle>
    <v-btn
      icon
      href="https://uptrace.dev/docs/querying-metrics.html#dashboards"
      target="_blank"
      class="ml-1"
    >
      <v-icon>mdi-help-circle-outline</v-icon>
    </v-btn>
  </span>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useDashManager, UseDashboard } from '@/metrics/use-dashboards'

export default defineComponent({
  name: 'DashToggleView',

  props: {
    dashboard: {
      type: Object as PropType<UseDashboard>,
      required: true,
    },
  },

  setup(props, ctx) {
    const dashMan = useDashManager()

    const view = computed({
      get(): string {
        return props.dashboard.data?.isTable ? 'table' : 'grid'
      },
      set(view: string) {
        dashMan.update({ isTable: view === 'table' }).then(() => {
          props.dashboard.reload().then(() => {
            ctx.emit('input:view', view)
          })
        })
      },
    })

    return { dashMan, view }
  },
})
</script>

<style lang="scss" scoped></style>
