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
            :systems="[system]"
            :loading="groups.loading"
            :groups="groups.items"
            :columns="groups.columns"
            :plottable-columns="groups.plottableColumns"
            :plotted-columns="[AttrKey.spanCountPerMin]"
            :order="groups.order"
            :events-mode="eventsMode"
            :axios-params="groups.axiosParams"
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
import { createUqlEditor, useQueryStore } from '@/use/uql'
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
    const { where } = useQueryStore()

    const system = computed(() => {
      return route.value.params.system
    })

    const eventsMode = computed(() => {
      return isEventSystem(system.value)
    })

    const query = computed(() => {
      return createUqlEditor()
        .exploreAttr(AttrKey.spanGroupId, eventsMode.value)
        .add(where.value)
        .toString()
    })

    const groups = useGroups(() => {
      return {
        ...props.axiosParams,
        system: system.value,
        query: query.value,
      }
    })
    groups.order.syncQueryParams()

    const groupListRoute = computed(() => {
      return {
        name: 'SpanGroupList',
        query: {
          ...groups.order.queryParams(),
          system: system.value,
          query: query.value,
        },
      }
    })

    return {
      AttrKey,
      system,
      eventsMode,
      groups,
      groupListRoute,
    }
  },
})
</script>

<style lang="scss" scoped></style>
