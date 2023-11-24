<template>
  <div>
    <TracingPlaceholder v-if="systems.dataHint" :date-range="dateRange" :systems="systems" />

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
        <DateRangePicker :date-range="dateRange" :range-days="90" />
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
              <v-tab :to="{ name: 'ServiceGraph', query: pick($route.query, 'system', 'query') }">
                Service graph
              </v-tab>
              <v-tab
                v-for="system in chosenSystems"
                :key="system.name"
                :to="{
                  name: 'SystemGroupList',
                  params: { system: system.name },
                  query: pick($route.query, 'system', 'query'),
                }"
              >
                {{ system.name }} ({{ system.groupCount }})
              </v-tab>
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
            <router-view name="overview" :date-range="dateRange" :systems="systems" :uql="uql" />
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
import { useAnnotations } from '@/org/use-annotations'
import { useTitle } from '@vueuse/core'
import { UseDateRange } from '@/use/date-range'
import { useUql, useQueryStore, provideQueryStore } from '@/use/uql'
import { useProject } from '@/org/use-projects'
import { useSystems, addAllSystem } from '@/tracing/system/use-systems'

// Components
import TracingPlaceholder from '@/tracing/TracingPlaceholder.vue'
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import SystemPicker from '@/tracing/system/SystemPicker.vue'
import QuickSpanFilter from '@/tracing/query/QuickSpanFilter.vue'
import SystemQuickMetrics from '@/tracing/system/SystemQuickMetrics.vue'

// Utilities
import { isSpanSystem, isErrorSystem, SystemName, AttrKey } from '@/models/otel'
import { DAY } from '@/util/fmt/date'

interface ChosenSystem {
  name: string
  groupCount: number
}

export default defineComponent({
  name: 'Overview',
  components: {
    TracingPlaceholder,
    DateRangePicker,
    SystemPicker,
    QuickSpanFilter,
    SystemQuickMetrics,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
  },

  setup(props) {
    useTitle('Overview')

    const project = useProject()

    const uql = useUql()
    provideQueryStore(useQueryStore(uql))

    useAnnotations(() => {
      return {
        ...props.dateRange.axiosParams(),
      }
    })

    const systems = useSystems(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...uql.axiosParams(),
      }
    })

    const spanSystems = computed(() => {
      const items = systems.items.filter((item) => isSpanSystem(item.system))
      addAllSystem(items, SystemName.SpansAll)
      return items
    })

    const chosenSystems = computed((): ChosenSystem[] => {
      if (props.dateRange.duration > 3 * DAY) {
        return []
      }

      const chosenMap = new Map<string, ChosenSystem>()

      for (let system of systems.items) {
        if (system.isGroup) {
          continue
        }
        if (!isErrorSystem(system.system)) {
          continue
        }

        const found = chosenMap.get(system.system)
        if (found) {
          found.groupCount += system.groupCount
        } else {
          chosenMap.set(system.system, {
            name: system.system,
            groupCount: system.groupCount,
          })
        }
      }

      return Array.from(chosenMap.values())
    })

    return {
      AttrKey,

      uql,
      project,
      systems,
      spanSystems,

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
