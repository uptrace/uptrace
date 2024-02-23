<template>
  <div>
    <portal to="navigation">
      <SystemGroupPicker
        :loading="systems.loading"
        :value="systems.activeSystems"
        :systems="systems.items"
        @update:systems="systemItems = $event"
      />
    </portal>

    <TracingPlaceholder v-if="systems.dataHint" :date-range="dateRange" :systems="systems" />

    <template v-else>
      <div class="border-bottom">
        <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="pb-0">
          <v-row align="center" class="mb-4">
            <v-col cols="auto">
              <SystemPicker
                v-model="systems.activeSystems"
                :loading="systems.loading"
                :systems="systemItems"
              />
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
                <v-tab :to="routes.spanList" exact-path>{{ spanListName }}</v-tab>
                <v-tab :to="routes.timeseries" exact-path>Timeseries</v-tab>
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
            :agg-disabled="$route.name === 'SpanList'"
            @click:reset="resetQuery(true)"
          />
        </UptraceQuery>

        <router-view
          name="tracing"
          :date-range="dateRange"
          :systems="systems"
          :uql="uql"
          :axios-params="axiosParams"
          :search-input.sync="searchInput"
        >
          <template slot="search-filter">
            <QuickSearch v-model="searchInput" />
          </template>
        </router-view>
      </v-container>
    </template>
  </div>
</template>

<script lang="ts">
import { pick } from 'lodash-es'
import { defineComponent, shallowRef, computed, proxyRefs, PropType } from 'vue'
import { refDebounced } from '@vueuse/core'

// Composables
import { useTitle } from '@vueuse/core'
import { useRoute } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { useUser } from '@/org/use-users'
import { useSystems, System } from '@/tracing/system/use-systems'
import { useUql, createQueryEditor, provideQueryStore, useQueryStore } from '@/use/uql'

// Components
import TracingPlaceholder from '@/tracing/TracingPlaceholder.vue'
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import SystemPicker from '@/tracing/system/SystemPicker.vue'
import SystemGroupPicker from '@/tracing/system/SystemGroupPicker.vue'
import SavedViews from '@/tracing/views/SavedViews.vue'
import UptraceQuery from '@/components/UptraceQuery.vue'
import SpanQueryBuilder from '@/tracing/query/SpanQueryBuilder.vue'
import QuickSearch from '@/components/QuickSearch.vue'

// Misc
import { SystemName, AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'Tracing',
  components: {
    TracingPlaceholder,
    DateRangePicker,
    SystemPicker,
    SystemGroupPicker,
    SavedViews,
    UptraceQuery,
    SpanQueryBuilder,
    QuickSearch,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
  },

  setup(props) {
    useTitle('Explore spans')

    const user = useUser()

    const searchInput = shallowRef('')
    const debouncedSearchInput = refDebounced(searchInput, 600)

    const uql = useUql()
    provideQueryStore(useQueryStore(uql))

    const systems = useSystems(() => {
      return {
        ...props.dateRange.axiosParams(),
        query: uql.whereQuery,
      }
    })

    const systemItems = shallowRef<System[]>([])

    const axiosParams = computed(() => {
      const params: Record<string, any> = {
        ...props.dateRange.axiosParams(),
        ...systems.axiosParams(),
        ...uql.axiosParams(),
      }
      if (debouncedSearchInput.value) {
        params.search = debouncedSearchInput.value
      }
      return params
    })

    const spanListName = computed(() => {
      switch (systems.groupName) {
        case SystemName.LogAll:
          return 'Logs'
        case SystemName.EventsAll:
          return 'Events'
        default:
          return 'Spans'
      }
    })

    function resetQuery(clear = false) {
      uql.query = createQueryEditor()
        .exploreAttr(AttrKey.spanGroupId, systems.isSpan)
        .add(clear ? '' : uql.whereQuery)
        .toString()
    }

    return {
      user,
      systems,
      systemItems,

      uql,
      axiosParams,
      searchInput,
      spanListName,

      routes: useRoutes(),

      resetQuery,
    }
  },
})

function useRoutes() {
  const route = useRoute()

  const groupList = computed(() => {
    return routeFor('SpanGroupList')
  })

  const spanList = computed(() => {
    return routeFor('SpanList')
  })

  const timeseries = computed(() => {
    return routeFor('SpanTimeseries')
  })

  function routeFor(routeName: string) {
    return {
      name: routeName,
      query: pick(route.value.query, ['system', 'query', 'time_gte', 'time_dur']),
    }
  }

  return proxyRefs({
    groupList,
    spanList,
    timeseries,
  })
}
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
