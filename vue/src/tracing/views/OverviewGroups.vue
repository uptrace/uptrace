<template>
  <v-card outlined rounded="lg">
    <v-toolbar flat color="light-blue lighten-5">
      <v-toolbar-title>{{ system }} groups</v-toolbar-title>
      <v-spacer />
      <v-btn :to="groupListRoute" small class="primary">View groups</v-btn>
    </v-toolbar>

    <v-container fluid>
      <v-row>
        <v-col>
          <GroupsList
            :date-range="dateRange"
            :loading="groups.loading"
            :is-resolved="groups.status.isResolved()"
            :groups="groups.items"
            :columns="groups.columns"
            :plottable-columns="groups.plottableColumns"
            :order="groups.order"
            :events-mode="eventsMode"
            :axios-params="internalAxiosParams"
          />
        </v-col>
      </v-row>
    </v-container>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { createUqlEditor } from '@/use/uql'
import { useGroups } from '@/tracing/use-explore-spans'

// Components
import GroupsList from '@/tracing/GroupsList.vue'

// Utilities
import { AttrKey, isEventSystem } from '@/models/otel'

export default defineComponent({
  name: 'OverviewGroups',
  components: { GroupsList },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    axiosParams: {
      type: Object,
      required: true,
    },
  },

  setup(props) {
    const route = useRoute()

    const system = computed(() => {
      return route.value.params.system
    })

    const eventsMode = computed(() => {
      return isEventSystem(system.value)
    })

    const internalAxiosParams = computed(() => {
      const ss = [
        createUqlEditor().exploreAttr(AttrKey.spanGroupId, eventsMode.value).toString(),
        route.value.query.query,
      ]
      return {
        ...props.axiosParams,
        system: system.value,
        query: ss.filter((v) => v).join(' | '),
      }
    })

    const groups = useGroups(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/tracing/${projectId}/groups`,
        params: internalAxiosParams.value,
      }
    })

    const groupListRoute = computed(() => {
      return {
        name: eventsMode.value ? 'EventGroupList' : 'SpanGroupList',
        query: {
          ...route.value.query,
          ...groups.order.queryParams(),
          system: system.value,
          query: createUqlEditor().exploreAttr(AttrKey.spanGroupId, eventsMode.value).toString(),
        },
      }
    })

    return {
      system,
      eventsMode,
      internalAxiosParams,
      groups,
      groupListRoute,
    }
  },
})
</script>

<style lang="scss" scoped></style>
