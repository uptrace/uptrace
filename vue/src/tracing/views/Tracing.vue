<template>
  <div>
    <HelpCard v-if="systems.hasNoData" :date-range="dateRange" :loading="systems.loading" />

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
            <v-col cols="auto">
              <SystemGroupPicker
                :loading="systems.loading"
                :systems="systems.items"
                :system="systems.activeSystems"
                @update:systems="systemItems = $event"
              />
            </v-col>

            <v-spacer />

            <v-col cols="auto">
              <DateRangePicker :date-range="dateRange" sync-query-params />
            </v-col>
          </v-row>

          <v-row align="end" no-gutters>
            <v-col cols="auto">
              <v-tabs :key="$route.fullPath" background-color="transparent">
                <v-tab :to="routes.groupList" exact-path>Groups</v-tab>
                <v-tab :to="routes.itemList" exact-path>Spans</v-tab>
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
            :agg-disabled="['SpanGroupList'].indexOf($route.name) === -1"
            @click:reset="resetQuery"
          />
        </UptraceQuery>

        <router-view
          name="tracing"
          :date-range="dateRange"
          :systems="systems"
          :uql="uql"
          :axios-params="axiosParams"
        />
      </v-container>
    </template>
  </div>
</template>

<script lang="ts">
import { pick } from 'lodash-es'
import { defineComponent, shallowRef, computed, watch, proxyRefs, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useTitle } from '@vueuse/core'
import { UseDateRange } from '@/use/date-range'
import { useUser } from '@/org/use-users'
import { useSystems, System } from '@/tracing/system/use-systems'
import { useUql, createUqlEditor, useProvideQueryStore } from '@/use/uql'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import SystemPicker from '@/tracing/system/SystemPicker.vue'
import SystemGroupPicker from '@/tracing/system/SystemGroupPicker.vue'
import HelpCard from '@/tracing/HelpCard.vue'
import SavedViews from '@/tracing/views/SavedViews.vue'
import UptraceQuery from '@/components/UptraceQuery.vue'
import SpanQueryBuilder from '@/tracing/query/SpanQueryBuilder.vue'

// Utilities
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'Tracing',
  components: {
    DateRangePicker,
    SystemPicker,
    SystemGroupPicker,
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
  },

  setup(props) {
    useTitle('Explore spans')

    const route = useRoute()
    const user = useUser()

    const uql = useUql()
    uql.syncQueryParams()
    useProvideQueryStore(uql)

    const systems = useSystems(() => {
      return {
        ...props.dateRange.axiosParams(),
        query: uql.whereQuery,
      }
    })
    systems.syncQueryParams()

    const systemItems = shallowRef<System[]>([])

    const axiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...uql.axiosParams(),
        system: systems.activeSystems,
      }
    })

    watch(
      () => systems.activeSystems,
      (activeSystem) => {
        if (activeSystem.length && !route.value.query.query) {
          resetQuery()
        }
      },
      { immediate: true, flush: 'post' },
    )

    function resetQuery(clear = false) {
      uql.query = createUqlEditor()
        .exploreAttr(AttrKey.spanGroupId, systems.isEvent)
        .add(clear ? '' : uql.whereQuery)
        .toString()
    }

    return {
      user,
      systems,
      systemItems,

      uql,
      axiosParams,

      routes: useRoutes(),

      resetQuery,
    }
  },
})

function useRoutes() {
  const route = useRoute()

  const groupList = computed(() => {
    return routeTo('SpanGroupList')
  })

  const itemList = computed(() => {
    return routeTo('SpanList')
  })

  function routeTo(routeName: string) {
    return {
      name: routeName,
      query: pick(route.value.query, ['system', 'query', 'time_gte', 'time_dur']),
    }
  }

  return proxyRefs({
    groupList,
    itemList,
  })
}
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
