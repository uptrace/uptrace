<template>
  <XPlaceholder>
    <template v-if="systems.hasNoData" #placeholder>
      <HelpCard :date-range="dateRange" :loading="systems.loading" />
    </template>

    <PageToolbar :loading="systems.loading" :fluid="$vuetify.breakpoint.mdAndDown">
      <QuickSpanFilter
        v-model="envs.active"
        :loading="envs.loading"
        :items="envs.items"
        :attr="AttrKey.serviceName"
        param-name="env"
      />
      <QuickSpanFilter
        v-model="services.active"
        :loading="services.loading"
        :items="services.items"
        :attr="AttrKey.deploymentEnvironment"
        param-name="service"
      />
      <v-spacer />
      <DateRangePicker :date-range="dateRange" />
    </PageToolbar>

    <div class="border">
      <div class="grey lighten-5">
        <v-container fluid class="mb-2">
          <SystemQuickMetrics :loading="systems.loading" :systems="systems.items" />
        </v-container>

        <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="pb-0">
          <v-tabs background-color="transparent">
            <v-tab :to="{ name: 'Overview' }">Systems</v-tab>
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

    <v-container :fluid="$vuetify.breakpoint.mdAndDown">
      <v-row>
        <v-col>
          <router-view
            :date-range="dateRange"
            :envs="envs"
            :services="services"
            :systems="systems"
          />
        </v-col>
      </v-row>
    </v-container>
  </XPlaceholder>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import type { UseDateRange } from '@/use/date-range'
import { useEnvs, useServices } from '@/tracing/use-quick-span-filters'
import { useProject } from '@/use/project'
import { useSystems } from '@/tracing/use-systems'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import QuickSpanFilter from '@/tracing/QuickSpanFilter.vue'
import HelpCard from '@/tracing/HelpCard.vue'
import SystemQuickMetrics from '@/tracing/overview/SystemQuickMetrics.vue'

// Utilities
import { AttrKey, SystemName } from '@/models/otelattr'
import { day } from '@/util/fmt/date'

export default defineComponent({
  name: 'Overview',
  components: { DateRangePicker, QuickSpanFilter, HelpCard, SystemQuickMetrics },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
  },

  setup(props) {
    useTitle('Overview')

    props.dateRange.syncQuery()

    const envs = useEnvs(props.dateRange)
    const services = useServices(props.dateRange)

    const project = useProject()
    const systems = useSystems(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...envs.axiosParams(),
        ...services.axiosParams(),
      }
    })

    const chosenSystems = computed((): string[] => {
      if (props.dateRange.duration > 3 * day) {
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
      project,
      envs,
      services,
      systems,
      chosenSystems,
    }
  },
})
</script>

<style lang="scss" scoped>
.border {
  overflow: auto;
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
