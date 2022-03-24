<template>
  <XPlaceholder>
    <template v-if="systems.hasNoData" #placeholder>
      <HelpCard :date-range="dateRange" :loading="systems.loading" />
    </template>

    <PageToolbar :loading="systems.loading">
      <v-spacer />
      <DateRangePicker :date-range="dateRange" />
    </PageToolbar>

    <div class="border">
      <div class="grey lighten-5">
        <v-container fluid class="mb-2">
          <SystemQuickMetrics :loading="systems.loading" :systems="systems.list" />
        </v-container>

        <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="pb-0">
          <v-tabs background-color="transparent">
            <v-tab :to="{ name: 'Overview' }">Systems</v-tab>
            <v-tab :to="{ name: 'ServiceOverview' }">Services</v-tab>
            <v-tab :to="{ name: 'HostOverview' }">Hosts</v-tab>
            <v-tab
              v-for="system in chosenSystems"
              :key="system"
              :to="{ name: 'SystemGroupList', params: { system: system } }"
            >
              {{ system }}
            </v-tab>
            <v-tab :to="{ name: 'SlowestGroups' }">Slowest groups</v-tab>
          </v-tabs>
        </v-container>
      </div>
    </div>

    <v-container :fluid="$vuetify.breakpoint.mdAndDown">
      <v-row>
        <v-col>
          <router-view :date-range="dateRange" :systems="systems" />
        </v-col>
      </v-row>
    </v-container>
  </XPlaceholder>
</template>

<script lang="ts">
import { defineComponent, computed } from '@vue/composition-api'

// Composables
import { useTitle } from '@vueuse/core'
import { useDateRange } from '@/use/date-range'
import { useSystems } from '@/use/systems'

// Components
import DateRangePicker from '@/components/DateRangePicker.vue'
import HelpCard from '@/components/HelpCard.vue'
import SystemQuickMetrics from '@/components/SystemQuickMetrics.vue'

// Utilities
import { xsys } from '@/models/otelattr'
import { day } from '@/util/date'

export default defineComponent({
  name: 'Overview',
  components: { DateRangePicker, HelpCard, SystemQuickMetrics },

  setup() {
    useTitle('Overview')

    const dateRange = useDateRange()
    dateRange.syncQuery()

    const systems = useSystems(dateRange)

    const chosenSystems = computed((): string[] => {
      if (dateRange.duration > 3 * day) {
        return []
      }

      const candidates = [xsys.logFatal, xsys.logPanic, xsys.logError, xsys.logWarn]
      const chosen = []
      for (let candidate of candidates) {
        const found = systems.list.find((v) => v.system === candidate)
        if (found) {
          chosen.push(candidate)
        }
      }
      return chosen
    })

    return {
      dateRange,
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
