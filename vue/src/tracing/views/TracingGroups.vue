<template>
  <XPlaceholder>
    <template v-if="explore.errorCode === 'invalid_query'" #placeholder>
      <v-row>
        <v-col>
          <v-banner>
            <v-icon slot="icon" color="error" size="36">mdi-alert-circle</v-icon>
            <span class="subtitle-1 text--secondary">{{ explore.errorMessage }}</span>
          </v-banner>

          <PrismCode v-if="explore.query" :code="explore.query" language="sql" />
        </v-col>
      </v-row>
    </template>

    <UptraceQuery :uql="uql" class="mt-1 mb-3">
      <SpanQueryBuilder
        :uql="uql"
        :systems="systems"
        :axios-params="axiosParams"
        :agg-disabled="['EventGroupList', 'SpanGroupList'].indexOf($route.name) === -1"
        @click:reset="resetQuery"
      />
    </UptraceQuery>

    <v-row>
      <v-col>
        <v-card rounded="lg" outlined class="mb-4">
          <v-toolbar flat color="light-blue lighten-5">
            <v-toolbar-title>
              <span>Groups</span>
            </v-toolbar-title>

            <v-spacer />

            <div class="text-body-2 blue-grey--text text--darken-3">
              <strong><XNum :value="explore.pager.numItem" verbose /></strong> groups
            </div>
          </v-toolbar>

          <v-card-text>
            <v-slide-group v-model="activeColumns" multiple center-active show-arrows class="mb-4">
              <v-slide-item
                v-for="(col, i) in explore.plotColumns"
                v-slot="{ active, toggle }"
                :key="col.name"
                :value="col.name"
              >
                <v-btn
                  :input-value="active"
                  active-class="blue white--text"
                  small
                  depressed
                  rounded
                  :class="{ 'ml-1': i > 0 }"
                  style="text-transform: none"
                  @click="toggle"
                >
                  {{ col.name }}
                </v-btn>
              </v-slide-item>
            </v-slide-group>

            <GroupsTable
              :date-range="dateRange"
              :events-mode="eventsMode"
              :uql="uql"
              :loading="explore.loading"
              :items="explore.pageItems"
              :columns="explore.columns"
              :group-columns="explore.groupColumns"
              :plot-columns="activeColumns"
              :order="explore.order"
              :axios-params="axiosParams"
              :show-system="showSystem"
            />
          </v-card-text>
        </v-card>

        <XPagination :pager="explore.pager" />
      </v-col>
    </v-row>
  </XPlaceholder>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseSystems } from '@/tracing/system/use-systems'
import { UseUql } from '@/use/uql'
import { useSpanExplore } from '@/tracing/use-span-explore'

// Components
import UptraceQuery from '@/components/UptraceQuery.vue'
import SpanQueryBuilder from '@/tracing/query/SpanQueryBuilder.vue'
import GroupsTable from '@/tracing/GroupsTable.vue'

import { isDummySystem } from '@/models/otel'

export default defineComponent({
  name: 'TracingGroups',
  components: { UptraceQuery, SpanQueryBuilder, GroupsTable },

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
    eventsMode: {
      type: Boolean,
      required: true,
    },
    query: {
      type: String,
      required: true,
    },
    spanListRoute: {
      type: String,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()
    const activeColumns = shallowRef<string[]>([])

    const explore = useSpanExplore(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/tracing/${projectId}/groups`,
        params: props.axiosParams,
      }
    })

    const showSystem = computed(() => {
      if (route.value.params.eventSystem) {
        return false
      }

      const systems = props.systems.activeSystem
      if (systems.length > 1) {
        return true
      }
      if (systems.length === 1) {
        return isDummySystem(systems[0])
      }
      return false
    })

    watch(
      () => explore.plotColumns,
      (allColumns) => {
        if (allColumns.length && !activeColumns.value.length) {
          activeColumns.value = [allColumns[0].name]
        }
      },
    )

    watch(
      () => explore.queryParts,
      (queryParts) => {
        if (queryParts) {
          props.uql.syncParts(queryParts)
        }
      },
    )

    watch(
      () => props.query,
      () => {
        if (!route.value.query.query) {
          resetQuery()
        }
      },
      { immediate: true },
    )

    function resetQuery() {
      props.uql.query = props.query
    }

    return {
      route,
      activeColumns,
      explore,
      showSystem,

      resetQuery,
    }
  },
})
</script>

<style lang="scss" scoped></style>
