<template>
  <v-card outlined rounded="lg">
    <v-toolbar flat color="light-blue lighten-5">
      <v-toolbar-title>{{ system }} groups</v-toolbar-title>
      <v-spacer />
      <v-btn :to="exploreRoute" small class="primary">View in explorer</v-btn>
    </v-toolbar>

    <v-card-text>
      <GroupsTable
        :date-range="dateRange"
        :systems="systems"
        :loading="explore.loading"
        :items="explore.pageItems"
        :columns="explore.columns"
        :group-columns="explore.groupColumns"
        :plot-columns="plottableColumns"
        :order="explore.order"
        :axios-params="axiosParams"
      />
    </v-card-text>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseSystems } from '@/tracing/use-systems'
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseEnvs, UseServices } from '@/tracing/use-quick-span-filters'
import { exploreAttr } from '@/use/uql'
import { useSpanExplore } from '@/tracing/use-span-explore'

// Components
import GroupsTable from '@/tracing/GroupsTable.vue'

// Utilities
import { AttrKey, isEventSystem } from '@/models/otelattr'

export default defineComponent({
  name: 'SystemGroupList',
  components: { GroupsTable },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    envs: {
      type: Object as PropType<UseEnvs>,
      required: true,
    },
    services: {
      type: Object as PropType<UseServices>,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const plottableColumns = computed(() => {
      return [AttrKey.spanCountPerMin]
    })

    const system = computed(() => {
      return route.value.params.system
    })

    const axiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...props.envs.axiosParams(),
        ...props.services.axiosParams(),
        system: system.value,
        query: exploreAttr(AttrKey.spanGroupId, isEventSystem(system.value)),
      }
    })

    const explore = useSpanExplore(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/tracing/${projectId}/groups`,
        params: axiosParams.value,
      }
    })

    const exploreRoute = computed(() => {
      return {
        name: isEventSystem(system.value) ? 'EventGroupList' : 'SpanGroupList',
        query: {
          ...route.value.query,
          ...explore.order.axiosParams,
          system: system.value,
          query: exploreAttr(AttrKey.spanGroupId, isEventSystem(system.value)),
        },
      }
    })

    return {
      plottableColumns,

      system,
      axiosParams,
      explore,
      exploreRoute,
    }
  },
})
</script>

<style lang="scss" scoped></style>
