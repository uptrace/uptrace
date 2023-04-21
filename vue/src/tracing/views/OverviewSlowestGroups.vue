<template>
  <v-card outlined rounded="lg">
    <v-toolbar flat color="light-blue lighten-5">
      <v-toolbar-title>Slowest groups</v-toolbar-title>
      <v-spacer />
      <v-btn :to="groupListRoute" small class="primary">View groups</v-btn>
    </v-toolbar>

    <v-container fluid>
      <GroupsList
        :date-range="dateRange"
        :loading="groups.loading"
        :is-resolved="groups.status.isResolved()"
        :groups="groups.items"
        :columns="groups.columns"
        :plottable-columns="groups.plottableColumns"
        :plotted-columns="plottedColumns"
        show-plotted-column-items
        :order="groups.order"
        :axios-params="internalAxiosParams"
        show-system
      />
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
import { SystemName, AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'OverviewSlowestGroups',
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
    const query = createUqlEditor().exploreAttr(AttrKey.spanGroupId).toString()

    const internalAxiosParams = computed(() => {
      return {
        ...props.axiosParams,
        query: [query, route.value.query.query].filter((v) => v).join(' | '),
      }
    })

    const groups = useGroups(
      () => {
        const { projectId } = route.value.params
        return {
          url: `/api/v1/tracing/${projectId}/groups`,
          params: internalAxiosParams.value,
        }
      },
      {
        order: {
          column: `p50(${AttrKey.spanDuration})`,
          desc: true,
        },
      },
    )

    const plottedColumns = [AttrKey.spanCountPerMin, `p50(${AttrKey.spanDuration})`]
    const groupListRoute = computed(() => {
      return {
        name: 'SpanGroupList',
        query: {
          ...route.value.query,
          ...groups.order.queryParams(),
          query,
        },
      }
    })

    return {
      SystemName,
      AttrKey,

      internalAxiosParams,
      groups,

      plottedColumns,
      groupListRoute,
    }
  },
})
</script>

<style lang="scss" scoped></style>
