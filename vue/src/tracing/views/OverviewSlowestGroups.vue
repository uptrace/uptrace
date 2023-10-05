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
        :systems="systems.activeSystems"
        :loading="groups.loading"
        :groups="groups.items"
        :columns="groups.columns"
        :plottable-columns="groups.plottableColumns"
        :plotted-columns="plottedColumns"
        show-plotted-column-items
        :order="groups.order"
        show-system
        :axios-params="groups.axiosParams"
      />
    </v-container>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { createUqlEditor, useQueryStore, provideQueryStore } from '@/use/uql'
import { useGroups } from '@/tracing/use-explore-spans'
import { UseSystems } from '@/tracing/system/use-systems'

// Components
import GroupsList from '@/tracing/GroupsList.vue'

// Utilities
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'OverviewSlowestGroups',
  components: { GroupsList },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
  },

  setup(props) {
    const { where } = useQueryStore()

    const query = computed(() => {
      return createUqlEditor().exploreAttr(AttrKey.spanGroupId).add(where.value).toString()
    })
    provideQueryStore({ query: computed(() => ''), where })

    const groups = useGroups(
      () => {
        return {
          ...props.dateRange.axiosParams(),
          ...props.systems.axiosParams(),
          query: query.value,
        }
      },
      {
        order: {
          column: `p50(${AttrKey.spanDuration})`,
          desc: true,
        },
      },
    )
    groups.order.syncQueryParams()

    const plottedColumns = [AttrKey.spanCountPerMin, `p50(${AttrKey.spanDuration})`]
    const groupListRoute = computed(() => {
      return {
        name: 'SpanGroupList',
        query: {
          ...groups.order.queryParams(),
          system: props.systems,
          query: query.value,
        },
      }
    })

    return {
      AttrKey,

      groups,

      plottedColumns,
      groupListRoute,
    }
  },
})
</script>

<style lang="scss" scoped></style>
