<template>
  <div>
    <template v-if="systems.hasNoData">
      <HelpCard :date-range="dateRange" :loading="systems.loading" />
    </template>

    <template v-else>
      <PageToolbar :loading="systems.loading" :fluid="$vuetify.breakpoint.lgAndDown">
        <SystemPicker
          v-if="systems.items.length"
          v-model="systems.activeSystem"
          :items="systems.items"
          all-system="all"
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
              <v-tab :to="{ name: 'SystemOverview' }">Systems</v-tab>
              <v-tab
                v-for="system in chosenSystems"
                :key="system"
                :to="{ name: 'SystemGroupList', params: { system: system } }"
              >
                {{ system }}
              </v-tab>
              <v-tab :to="{ name: 'SlowestGroups' }">Slowest groups</v-tab>
              <v-tab
                v-for="attr in project.pinnedAttrs"
                :key="attr"
                :to="{ name: 'AttrOverview', params: { attr } }"
                >{{ attr }}</v-tab
              >
            </v-tabs>
          </v-container>
        </div>
      </div>

      <v-container :fluid="$vuetify.breakpoint.lgAndDown">
        <v-row>
          <v-col>
            <router-view :date-range="dateRange" :axios-params="axiosParams" />
          </v-col>
        </v-row>
      </v-container>
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { UseDateRange } from '@/use/date-range'
import { useUql } from '@/use/uql'
import { useProject } from '@/org/use-projects'
import { useSystems } from '@/tracing/system/use-systems'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import SystemPicker from '@/tracing/system/SystemPicker.vue'
import QuickSpanFilter from '@/tracing/query/QuickSpanFilter.vue'
import SystemQuickMetrics from '@/tracing/system/SystemQuickMetrics.vue'
import HelpCard from '@/tracing/HelpCard.vue'

// Utilities
import { AttrKey, SystemName } from '@/models/otel'
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

    const systems = useSystems(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...uql.axiosParams(),
      }
    })
    systems.syncQueryParams()

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

      const candidates = [
        SystemName.logFatal,
        SystemName.logPanic,
        SystemName.logError,
        SystemName.logWarn,
      ]
      const chosen = []
      for (let candidate of candidates) {
        const found = systems.items.find((v) => v.system === candidate)
        if (found) {
          chosen.push(candidate)
        }
      }
      return chosen
    })

    return {
      AttrKey,

      uql,
      project,
      systems,
      axiosParams,

      chosenSystems,
    }
  },
})
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
