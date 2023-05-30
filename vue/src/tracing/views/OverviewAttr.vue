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
            :loading="groups.loading"
            :is-resolved="groups.status.isResolved()"
            :groups="groups.items"
            :columns="groups.columns"
            :plottable-columns="groups.plottableColumns"
            :plotted-columns="plottedColumns"
            :order="groups.order"
            :axios-params="internalAxiosParams"
          />
        </v-col>
      </v-row>
    </v-container>
  </v-card>
</template>

<script lang="ts">
import { omit } from 'lodash-es'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { createUqlEditor } from '@/use/uql'
import { useGroups } from '@/tracing/use-explore-spans'

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
    axiosParams: {
      type: Object,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const attr = computed(() => {
      return route.value.params.attr ?? AttrKey.spanSystem
    })

    const query = computed(() => {
      return createUqlEditor()
        .exploreAttr(attr.value)
        .add(`max(${AttrKey.spanDuration})`)
        .toString()
    })

    const internalAxiosParams = computed(() => {
      const ss = [query.value, route.value.query.query]
      return {
        ...props.axiosParams,
        query: ss.join(' | '),
      }
    })

    const groups = useGroups(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/tracing/${projectId}/groups`,
        params: internalAxiosParams.value,
      }
    })

    const plottedColumns = computed(() => {
      return groups.plottableColumns
        .map((col) => col.name)
        .filter((colName) => colName !== `max(${AttrKey.spanDuration})`)
    })

    const groupListRoute = computed(() => {
      return {
        name: 'SpanGroupList',
        query: {
          ...omit(route.value.query, 'columns'),
          query: query.value,
        },
      }
    })

    const spanListRoute = computed(() => {
      return {
        name: 'SpanList',
        query: {
          ...omit(route.value.query, 'sort_by', 'sort_desc'),
          query: query.value,
        },
      }
    })

    const metric = computed((): string => {
      switch (attr.value) {
        case AttrKey.serviceName:
          return 'uptrace.tracing.services'
        case AttrKey.hostName:
          return 'uptrace.tracing.hosts'
        default:
          return ''
      }
    })

    return {
      AttrKey,

      attr,
      internalAxiosParams,
      groups,
      plottedColumns,
      groupListRoute,
      spanListRoute,
      metric,
    }
  },
})
</script>

<style lang="scss"></style>
