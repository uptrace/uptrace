<template>
  <div>
    <v-card outlined rounded="lg" class="mb-4">
      <v-toolbar flat color="light-blue lighten-5">
        <v-toolbar-title>Slowest groups</v-toolbar-title>
        <v-spacer />
        <v-btn :to="groupRoute" small class="primary">View in explorer</v-btn>
      </v-toolbar>

      <v-card-text>
        <GroupsTable
          :date-range="dateRange"
          :systems="systems"
          :uql="uql"
          :loading="explore.loading"
          :items="explore.pageItems"
          :columns="explore.columns"
          :group-columns="explore.groupColumns"
          :plot-columns="activeColumns"
          :order="explore.order"
          :axios-params="axiosParams"
        />
      </v-card-text>
    </v-card>

    <XPagination :pager="explore.pager" />
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, computed, PropType } from 'vue'

// Composables
import { UseSystems } from '@/use/systems'
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseEnvs, UseServices } from '@/tracing/use-sticky-filters'
import { exploreAttr } from '@/use/uql'
import { useSpanExplore } from '@/tracing/use-span-explore'
import { useUql } from '@/use/uql'

// Components
import GroupsTable from '@/tracing/GroupsTable.vue'

// Utilities
import { xkey, xsys } from '@/models/otelattr'

export default defineComponent({
  name: 'SlowestGroups',
  components: { GroupsTable },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
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
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()
    const system = xsys.allSpans
    const query = exploreAttr(xkey.spanGroupId)

    const activeColumns = shallowRef<string[]>([])

    const uql = useUql({
      query: exploreAttr(xkey.spanGroupId),
      syncQuery: true,
    })

    const axiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...props.envs.axiosParams(),
        ...props.services.axiosParams(),
        system: system,
        query,
      }
    })

    const explore = useSpanExplore(
      () => {
        const { projectId } = route.value.params
        return {
          url: `/api/v1/tracing/${projectId}/groups`,
          params: axiosParams.value,
        }
      },
      {
        order: {
          column: `p50(${xkey.spanDuration})`,
          desc: true,
        },
      },
    )

    const groupRoute = computed(() => {
      return {
        name: 'SpanGroupList',
        query: {
          ...route.value.query,
          ...explore.order.axiosParams, // ?
          system,
          query,
        },
      }
    })

    watch(
      () => explore.plotColumns,
      (allColumns) => {
        if (allColumns.length && !activeColumns.value.length) {
          activeColumns.value = [allColumns[0].name]
        }
      },
    )

    return {
      xkey,
      uql,

      axiosParams,
      system,
      explore,
      activeColumns,
      groupRoute,
    }
  },
})
</script>

<style lang="scss"></style>
