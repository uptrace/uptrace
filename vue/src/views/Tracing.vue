<template>
  <XPlaceholder>
    <template v-if="systems.hasNoData" #placeholder>
      <HelpCard :date-range="dateRange" :loading="systems.loading" />
    </template>

    <div class="border">
      <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="pb-0">
        <v-row align="center" justify="space-between" class="mb-4">
          <v-col cols="auto">
            <SystemPicker
              :date-range="dateRange"
              :systems="systems"
              :tree="systemTree"
              :route-name="groupListRoute"
            />
          </v-col>

          <v-col cols="auto">
            <DateRangePicker :date-range="dateRange" />
          </v-col>
        </v-row>

        <v-row align="end" no-gutters>
          <v-col>
            <v-tabs :key="$route.fullPath" background-color="transparent">
              <v-tab :to="routes.groupList" exact-path>Groups</v-tab>
              <v-tab :to="routes.spanList" exact-path>{{
                spanListRoute == 'SpanList' ? 'Spans' : 'Logs'
              }}</v-tab>
            </v-tabs>
          </v-col>
        </v-row>
      </v-container>
    </div>

    <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="pt-2">
      <UptraceQuery :uql="uql" class="mb-1">
        <SpanFilters
          :uql="uql"
          :systems="systems"
          :axios-params="axiosParams"
          :group-list-route="groupListRoute"
          @click:reset="resetQuery"
        />
      </UptraceQuery>

      <router-view
        :date-range="dateRange"
        :systems="systems"
        :uql="uql"
        :axios-params="axiosParams"
        :span-list-route="spanListRoute"
        :group-list-route="groupListRoute"
      />
    </v-container>
  </XPlaceholder>
</template>

<script lang="ts">
import { clone } from 'lodash'
import { defineComponent, computed, watch, proxyRefs, PropType } from '@vue/composition-api'

// Composables
import { useRouter } from '@/use/router'
import { useTitle } from '@vueuse/core'
import { useDateRange } from '@/use/date-range'
import { useSystems, buildSystemTree, SystemTree, SystemFilter } from '@/use/systems'
import { useUql } from '@/use/uql'

// Components
import DateRangePicker from '@/components/DateRangePicker.vue'
import SystemPicker from '@/components/SystemPicker.vue'
import HelpCard from '@/components/HelpCard.vue'
import UptraceQuery from '@/components/uql/UptraceQuery.vue'
import SpanFilters from '@/components/uql/SpanFilters.vue'

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
    UptraceQuery,
    SpanFilters,
  },

  props: {
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
    systemFilter: {
      type: Function as PropType<SystemFilter>,
      default: undefined,
    },
  },

  setup(props) {
    useTitle('Explore spans')

    const dateRange = useDateRange()
    dateRange.syncQuery()

    const systems = useSystems(dateRange)
    const uql = useUql({
      query: props.query,
      syncQuery: true,
    })

    const axiosParams = computed(() => {
      return {
        ...dateRange.axiosParams(),
        ...uql.axiosParams(),
        system: systems.activeValue,
      }
    })

    const systemTree = computed((): SystemTree[] => {
      let items = systems.list
      if (props.systemFilter) {
        items = items.filter(props.systemFilter)
      }
      return buildSystemTree(items)
    })

    watch(
      () => props.query,
      () => {
        resetQuery()
      },
      { immediate: true },
    )

    function resetQuery() {
      uql.query = props.query
    }

    return {
      dateRange,
      uql,
      axiosParams,
      systems,
      systemTree,
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
