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
            :systems="systems"
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
import { omit } from 'lodash-es'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { createUqlEditor, useQueryStore } from '@/use/uql'
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
    systems: {
      type: Array as PropType<string[]>,
      required: true,
    },
    axiosParams: {
      type: Object,
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

    const groups = useGroups(() => {
      return {
        ...props.axiosParams,
        query: query.value,
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
          system: props.systems,
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
      spanListRoute,
    }
  },
})
</script>

<style lang="scss"></style>
