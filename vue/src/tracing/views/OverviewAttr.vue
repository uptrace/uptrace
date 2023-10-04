<template>
  <v-card outlined rounded="lg">
    <v-toolbar flat color="light-blue lighten-5">
      <v-toolbar-title>{{ attr }} overview</v-toolbar-title>
      <v-spacer />
      <v-btn :to="groupListRoute" small class="primary">View groups</v-btn>
    </v-toolbar>

    <v-container fluid>
      <v-row>
        <v-col>
          <GroupsList
            :date-range="dateRange"
            :systems="systems.activeSystems"
            :loading="groups.loading"
            :groups="groups.items"
            :columns="groups.columns"
            :plottable-columns="groups.plottableColumns"
            :plotted-columns="plottedColumns"
            :order="groups.order"
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
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { createUqlEditor, useQueryStore, provideQueryStore } from '@/use/uql'
import { useGroups } from '@/tracing/use-explore-spans'
import { UseSystems } from '@/tracing/system/use-systems'

// Components
import GroupsList from '@/tracing/GroupsList.vue'

// Utilities
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'OverviewAttr',
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
    const { route } = useRouter()
    const { where } = useQueryStore()

    const attr = computed(() => {
      return route.value.params.attr ?? AttrKey.spanSystem
    })

    const query = computed(() => {
      return createUqlEditor()
        .exploreAttr(attr.value)
        .add(`max(${AttrKey.spanDuration})`)
        .add(where.value)
        .toString()
    })
    provideQueryStore({ query, where })

    const groups = useGroups(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...props.systems.axiosParams(),
        query: query.value,
      }
    })
    groups.order.syncQueryParams()

    const plottedColumns = computed(() => {
      return groups.plottableColumns
        .map((col) => col.name)
        .filter((colName) => colName !== `max(${AttrKey.spanDuration})`)
    })

    const groupListRoute = computed(() => {
      return {
        name: 'SpanGroupList',
        query: {
          ...groups.order.queryParams(),
          system: props.systems.activeSystems,
          query: query.value,
        },
      }
    })

    return {
      AttrKey,

      attr,
      groups,
      plottedColumns,
      groupListRoute,
    }
  },
})
</script>

<style lang="scss"></style>
