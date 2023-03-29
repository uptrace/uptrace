<template>
  <div>
    <HelpCard v-if="systems.hasNoData" :date-range="dateRange" :loading="systems.loading" />

    <template v-else>
      <div class="border-bottom">
        <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="pb-0">
          <v-row align="center" class="mb-4">
            <v-col cols="auto">
              <div class="mr-4">
                <SystemPicker
                  v-model="systems.activeSystem"
                  :loading="systems.loading"
                  :items="systemsItems"
                  :all-system="allSystem"
                />
              </div>
            </v-col>
            <v-col cols="auto">
              <v-btn-toggle mandatory group color="blue accent-3">
                <v-btn :to="{ name: 'SpanGroupList' }">Spans</v-btn>
                <v-btn :to="{ name: 'EventGroupList' }">Events</v-btn>
              </v-btn-toggle>
            </v-col>

            <v-spacer />

            <v-col cols="auto">
              <DateRangePicker :date-range="dateRange" />
            </v-col>
          </v-row>

          <v-row align="end" no-gutters>
            <v-col cols="auto">
              <v-tabs :key="$route.fullPath" background-color="transparent">
                <v-tab :to="routes.groupList" exact-path>Groups</v-tab>
                <v-tab :to="routes.spanList" exact-path>{{
                  $route.name.startsWith('Span') ? 'Spans' : 'Events'
                }}</v-tab>
              </v-tabs>
            </v-col>
            <v-col cols="auto" class="ml-16 align-self-center">
              <SavedViews />
            </v-col>
          </v-row>
        </v-container>
      </div>

      <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="pt-2">
        <UptraceQuery :uql="uql" class="mt-1 mb-3">
          <SpanQueryBuilder
            :uql="uql"
            :systems="systems"
            :axios-params="axiosParams"
            :agg-disabled="['EventGroupList', 'SpanGroupList'].indexOf($route.name) === -1"
            @click:reset="resetQuery"
          />
        </UptraceQuery>

        <router-view
          name="tracing"
          :date-range="dateRange"
          :systems="systems"
          :uql="uql"
          :events-mode="eventsMode"
          :axios-params="axiosParams"
        />
      </v-container>
    </template>
  </div>
</template>

<script lang="ts">
import { clone } from 'lodash-es'
import { defineComponent, computed, watch, proxyRefs, PropType } from 'vue'

// Composables
import { useRouter, useRoute } from '@/use/router'
import { useTitle } from '@vueuse/core'
import { UseDateRange } from '@/use/date-range'
import { useUser } from '@/org/use-users'
import { useSystems, SystemsFilter } from '@/tracing/system/use-systems'
import { useUql } from '@/use/uql'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import SystemPicker from '@/tracing/system/SystemPicker.vue'
import HelpCard from '@/tracing/HelpCard.vue'
import SavedViews from '@/tracing/views/SavedViews.vue'
import UptraceQuery from '@/components/UptraceQuery.vue'
import SpanQueryBuilder from '@/tracing/query/SpanQueryBuilder.vue'

interface Props {
  itemListRouteName: string
  groupListRouteName: string
}

export default defineComponent({
  name: 'Tracing',
  components: {
    DateRangePicker,
    SystemPicker,
    HelpCard,
    SavedViews,
    UptraceQuery,
    SpanQueryBuilder,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systemsFilter: {
      type: Function as PropType<SystemsFilter>,
      default: undefined,
    },
    allSystem: {
      type: String,
      required: true,
    },
    eventsMode: {
      type: Boolean,
      required: true,
    },
    defaultQuery: {
      type: String,
      required: true,
    },
    itemListRouteName: {
      type: String,
      required: true,
    },
    groupListRouteName: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    useTitle('Explore spans')
    props.dateRange.syncQueryParams()
    props.dateRange.roundUp()

    const route = useRoute()
    const user = useUser()

    const uql = useUql()
    uql.syncQueryParams()

    const systems = useSystems(() => {
      return {
        ...props.dateRange.axiosParams(),
        query: uql.whereQuery,
      }
    })
    systems.syncQueryParams()

    const axiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...uql.axiosParams(),
        system: systems.activeSystem,
      }
    })

    const systemsItems = computed(() => {
      if (props.systemsFilter) {
        return props.systemsFilter(systems.items)
      }
      return systems.items
    })

    watch(
      () => props.defaultQuery,
      () => {
        if (!route.value.query.query) {
          resetQuery()
        }
      },
      { immediate: true, flush: 'pre' },
    )

    function resetQuery() {
      uql.query = props.defaultQuery
    }

    return {
      user,
      systems,
      systemsItems,

      uql,
      axiosParams,

      routes: useRoutes(props),

      resetQuery,
    }
  },
})

function useRoutes(props: Props) {
  const { route } = useRouter()

  const spanList = computed(() => {
    const query = clone(route.value.query)
    if (query.sort_by) {
      delete query.sort_by
    }

    return {
      name: props.itemListRouteName,
      query,
    }
  })

  const groupList = computed(() => {
    return {
      name: props.groupListRouteName,
      query: route.value.query,
    }
  })

  return proxyRefs({
    groupList,
    spanList,
  })
}
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
