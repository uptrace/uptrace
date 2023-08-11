<template>
  <div>
    <template v-if="systems.hasNoData">
      <HelpCard :date-range="dateRange" :loading="systems.loading" />
    </template>

    <template v-else>
      <PageToolbar :loading="systems.loading" :fluid="$vuetify.breakpoint.lgAndDown">
        <v-toolbar-items>
          <SystemPicker
            v-if="systems.items.length"
            v-model="systems.activeSystems"
            :systems="spanSystems"
            outlined
          />
          <QuickSpanFilter
            :date-range="dateRange"
            :uql="uql"
            name="env"
            :attr-key="AttrKey.deploymentEnvironment"
            class="ml-2"
          />
          <QuickSpanFilter
            :date-range="dateRange"
            :uql="uql"
            name="service"
            :attr-key="AttrKey.serviceName"
            class="ml-2"
          />
        </v-toolbar-items>

        <v-spacer />

        <DateRangePicker :date-range="dateRange" />
      </PageToolbar>

      <div class="border-bottom">
        <div class="grey lighten-5">
          <v-container fluid class="mb-2">
            <SystemQuickMetrics :loading="systems.loading" :systems="systems.items" />
          </v-container>

          <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="pb-0">
            <v-tabs background-color="transparent">
              <v-tab :to="{ name: 'SystemOverview', query: pick($route.query, 'system', 'query') }"
                >Systems</v-tab
              >
              <v-tab
                v-for="system in chosenSystems"
                :key="system"
                :to="{
                  name: 'SystemGroupList',
                  params: { system },
                  query: pick($route.query, 'system', 'query'),
                }"
                >{{ system }}</v-tab
              >
              <v-tab :to="{ name: 'SlowestGroups', query: pick($route.query, 'system', 'query') }"
                >Slowest</v-tab
              >
              <v-tab
                v-for="attr in project.pinnedAttrs"
                :key="attr"
                :to="{
                  name: 'AttrOverview',
                  params: { attr },
                  query: pick($route.query, 'system', 'query'),
                }"
                >{{ attr }}</v-tab
              >
            </v-tabs>
          </v-container>
        </div>
      </div>

      <v-container :fluid="$vuetify.breakpoint.lgAndDown">
        <v-row>
          <v-col>
            <router-view
              :date-range="dateRange"
              :systems="systems.activeSystems"
              :axios-params="axiosParams"
            />
          </v-col>
        </v-row>
      </v-container>
    </template>
  </div>
</template>

<script lang="ts">
import { pick } from 'lodash-es'
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { UseDateRange } from '@/use/date-range'
import { useUql, useProvideQueryStore } from '@/use/uql'
import { useProject } from '@/org/use-projects'
import { useSystems, addAllSystem } from '@/tracing/system/use-systems'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import SystemPicker from '@/tracing/system/SystemPicker.vue'
import QuickSpanFilter from '@/tracing/query/QuickSpanFilter.vue'
import SystemQuickMetrics from '@/tracing/system/SystemQuickMetrics.vue'
import HelpCard from '@/tracing/HelpCard.vue'

// Utilities
import { isSpanSystem, isErrorSystem, SystemName, AttrKey } from '@/models/otel'
import { DAY } from '@/util/fmt/date'

export default defineComponent({
  name: 'Overview',
  components: { DateRangePicker, SystemPicker, QuickSpanFilter, HelpCard, SystemQuickMetrics },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
  },

  setup(props) {
    useTitle('Overview')
    props.dateRange.syncQueryParams()

    const project = useProject()

    const uql = useUql()
    uql.syncQueryParams()
    useProvideQueryStore(uql)

    const systems = useSystems(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...uql.axiosParams(),
      }
    })
    systems.syncQueryParams()

    const spanSystems = computed(() => {
      const items = systems.items.filter((item) => isSpanSystem(item.system))
      addAllSystem(items, SystemName.SpansAll)
      return items
    })

    const axiosParams = computed(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...uql.axiosParams(),
        ...systems.axiosParams(),
      }
    })

    const chosenSystems = computed((): string[] => {
      if (props.dateRange.duration > 3 * DAY) {
        return []
      }

      const chosen = []
      for (let system of systems.items) {
        if (isErrorSystem(system.system)) {
          chosen.push(system.system)
        }
      }
      return chosen
    })

    return {
      AttrKey,

      uql,
      project,
      systems,
      spanSystems,
      axiosParams,

      chosenSystems,
      pick,
    }
  },
})
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
