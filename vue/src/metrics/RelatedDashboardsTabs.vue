<template>
  <v-btn-toggle v-if="relatedDashboards.length > 1" group dense color="primary">
    <v-btn
      v-for="dash in relatedDashboards"
      :key="dash.id"
      :to="{ name: 'DashboardShow', params: { dashId: dash.id } }"
    >
      {{ dashName(dash.name) }}
    </v-btn>
  </v-btn-toggle>
</template>

<script lang="ts">
import { orderBy, upperFirst } from 'lodash-es'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { Dashboard } from '@/metrics/types'

export default defineComponent({
  name: 'RelatedDashboardsTabs',

  props: {
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
    dashboards: {
      type: Array as PropType<Dashboard[]>,
      required: true,
    },
  },

  setup(props) {
    const prefix = computed(() => {
      const name = props.dashboard.name
      const i = name.indexOf(': ')
      if (i === -1) {
        return ''
      }
      return name.slice(0, i + 2)
    })

    const relatedDashboards = computed(() => {
      if (!prefix.value) {
        return []
      }

      const related: Dashboard[] = []

      for (let dash of props.dashboards) {
        if (dash.name.startsWith(prefix.value)) {
          related.push(dash)
        }
      }

      return orderBy(related, 'templateId', 'asc')
    })

    function dashName(name: string) {
      name = name.slice(prefix.value.length)
      return upperFirst(name)
    }

    return { relatedDashboards, dashName }
  },
})
</script>

<style lang="scss" scoped></style>
