<template>
  <XPlaceholder>
    <template v-if="systems.hasNoData" #placeholder>
      <HelpCard :date-range="dateRange" :loading="systems.loading" />
    </template>

    <div class="border">
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
      <router-view
        :date-range="dateRange"
        :systems="systems"
        :uql="uql"
        :events-mode="eventsMode"
        :query="query"
        :span-list-route="spanListRoute"
        :group-list-route="groupListRoute"
        :axios-params="axiosParams"
      />
    </v-container>
  </XPlaceholder>
</template>

<script lang="ts">
import { clone } from 'lodash-es'
import { defineComponent, computed, proxyRefs, PropType } from 'vue'

// Composables
import { useRouter, useRoute } from '@/use/router'
import { useTitle } from '@vueuse/core'
import { UseDateRange } from '@/use/date-range'
import { useUser } from '@/use/org'
import { useSystems, SystemsFilter } from '@/tracing/system/use-systems'
import { useUql } from '@/use/uql'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import SystemPicker from '@/tracing/system/SystemPicker.vue'
import HelpCard from '@/tracing/HelpCard.vue'

interface Props {
  spanListRoute: string
  groupListRoute: string
}

export default defineComponent({
  name: 'Tracing',
  components: {
    DateRangePicker,
    SystemPicker,
    HelpCard,
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
    query: {
      type: String,
      required: true,
    },
    spanListRoute: {
      type: String,
      required: true,
    },
    groupListRoute: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    useTitle('Explore spans')
    props.dateRange.syncQuery()
    props.dateRange.roundUp()

    const route = useRoute()
    const user = useUser()

    const uql = useUql({
      syncQuery: true,
    })

    const systems = useSystems(() => {
      return {
        ...props.dateRange.axiosParams(),
        query: uql.whereQuery,
      }
    })
    systems.syncQuery()

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

    return {
      route,
      user,
      systems,
      systemsItems,
      uql,
      axiosParams,
      routes: useRoutes(props),
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
      name: props.spanListRoute,
      query,
    }
  })

  const groupList = computed(() => {
    return {
      name: props.groupListRoute,
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
.border {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
