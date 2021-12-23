<template>
  <XPlaceholder>
    <template v-if="systems.hasNoData" #placeholder>
      <HelpCard :date-range="dateRange" :loading="systems.loading" />
    </template>

    <div>
      <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="pb-0">
        <v-row align="center" class="mb-4">
          <v-spacer />

          <v-col cols="auto">
            <DateRangePicker :date-range="dateRange" />
          </v-col>
        </v-row>
      </v-container>
    </div>

    <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="pt-2">
      <v-row>
        <v-col>
          <SystemOverview :date-range="dateRange" :systems="systems" />
        </v-col>
      </v-row>
    </v-container>
  </XPlaceholder>
</template>

<script lang="ts">
import { defineComponent } from '@vue/composition-api'

// Composables
import { useTitle } from '@vueuse/core'
import { useDateRange } from '@/use/date-range'
import { useSystems } from '@/use/systems'

// Components
import DateRangePicker from '@/components/DateRangePicker.vue'
import HelpCard from '@/components/HelpCard.vue'
import SystemOverview from '@/components/SystemOverview.vue'

export default defineComponent({
  name: 'Overview',
  components: { DateRangePicker, HelpCard, SystemOverview },

  setup() {
    useTitle('Overview')
    const dateRange = useDateRange()
    const systems = useSystems(dateRange)

    return { dateRange, systems }
  },
})
</script>

<style lang="scss" scoped></style>
