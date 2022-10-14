<template>
  <XPlaceholder>
    <template v-if="systems.hasNoData" #placeholder>
      <HelpCard :date-range="dateRange" :loading="systems.loading" />
    </template>

    <div class="border">
      <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="pb-0">
        <v-row align="center" justify="space-between" class="mb-4">
          <v-col class="d-flex align-center">
            <div class="mr-4">
              <SystemPicker
                v-if="systemsItems.length"
                :date-range="dateRange"
                :systems="systems"
                :items="systemsItems"
                :all-system="allSystem"
              />
            </div>
            <StickyFilter
              v-if="envs.items.length > 1"
              v-model="envs.active"
              :loading="envs.loading"
              :items="envs.items"
              param-name="env"
            />
            <StickyFilter
              v-if="services.items.length > 1"
              v-model="services.active"
              :loading="services.loading"
              :items="services.items"
              param-name="service"
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
      <router-view
        :date-range="dateRange"
        :systems="systems"
        :envs="envs"
        :query="query"
        :span-list-route="spanListRoute"
        :group-list-route="groupListRoute"
      />
    </v-container>
  </XPlaceholder>
</template>

<script lang="ts">
import { clone } from 'lodash-es'
import { defineComponent, computed, proxyRefs, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useTitle } from '@vueuse/core'
import { UseDateRange } from '@/use/date-range'
import { useEnvs, useServices } from '@/tracing/use-sticky-filters'
import { useUser } from '@/use/org'
import { useSystems, SystemsFilter } from '@/use/systems'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import SystemPicker from '@/tracing/SystemPicker.vue'
import StickyFilter from '@/tracing/StickyFilter.vue'
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
    StickyFilter,
    HelpCard,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
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
    systemsFilter: {
      type: Function as PropType<SystemsFilter>,
      default: undefined,
    },
    allSystem: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    useTitle('Explore spans')
    props.dateRange.syncQuery()

    const user = useUser()
    const envs = useEnvs(props.dateRange)
    const services = useServices(props.dateRange)

    const systems = useSystems(() => {
      return {
        ...props.dateRange.axiosParams(),
      }
    })

    const systemsItems = computed(() => {
      if (props.systemsFilter) {
        return props.systemsFilter(systems.items)
      }
      return systems.items
    })

    return {
      envs,
      services,
      user,
      systems,
      systemsItems,
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
