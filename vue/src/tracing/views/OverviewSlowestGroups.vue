<template>
  <v-card outlined rounded="lg">
    <v-toolbar flat color="light-blue lighten-5">
      <v-toolbar-title>Slowest groups</v-toolbar-title>
      <v-spacer />
      <v-btn :to="groupListRoute" small class="primary">View groups</v-btn>
    </v-toolbar>

    <v-container fluid>
      <ApiErrorCard v-if="groups.error" :error="groups.error" />
      <PagedGroupsCard
        v-else
        :date-range="dateRange"
        :systems="systems.activeSystems"
        :loading="groups.loading"
        :groups="groups.items"
        :columns="groups.columns"
        :plottable-columns="groups.plottableColumns"
        :plotted-columns="plottedColumns"
        show-plotted-column-items
        :order="groups.order"
        :axios-params="groups.axiosParams"
      />
    </v-container>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useSyncQueryParams } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { createQueryEditor, injectQueryStore, provideQueryStore, UseUql } from '@/use/uql'
import { UseSystems } from '@/tracing/system/use-systems'
import { useGroups } from '@/tracing/use-explore-spans'

// Components
import ApiErrorCard from '@/components/ApiErrorCard.vue'
import PagedGroupsCard from '@/tracing/PagedGroupsCard.vue'

// Misc
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'OverviewSlowestGroups',
  components: { ApiErrorCard, PagedGroupsCard },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
  },

  setup(props) {
    const { where } = injectQueryStore()

    const query = computed(() => {
      return createQueryEditor().exploreAttr(AttrKey.spanGroupId, true).add(where.value).toString()
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

    useSyncQueryParams({
      fromQuery(queryParams) {
        queryParams.setDefault('sort_by', `p50(${AttrKey.spanDuration})`)
        queryParams.setDefault('sort_desc', true)

        props.dateRange.parseQueryParams(queryParams)
        props.systems.parseQueryParams(queryParams)
        props.uql.parseQueryParams(queryParams)
        groups.order.parseQueryParams(queryParams)
      },
      toQuery() {
        return {
          ...props.dateRange.queryParams(),
          ...props.systems.queryParams(),
          ...props.uql.queryParams(),
          ...groups.order.queryParams(),
        }
      },
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
