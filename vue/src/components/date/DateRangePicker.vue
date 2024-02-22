<template>
  <div class="d-flex align-center justify-end flex-wrap">
    <div class="d-flex align-center flex-nowrap">
      <v-btn
        icon
        :disabled="!dateRange.hasPrevPeriod"
        title="Previous period"
        @click="dateRange.prevPeriod"
      >
        <v-icon class="small">mdi-chevron-left</v-icon>
      </v-btn>

      <DateRangePickerMenu :date-range="dateRange" />
      <PeriodPickerMenu
        :value="dateRange.duration"
        :periods="periods"
        @input="dateRange.changeDuration($event)"
      />

      <v-btn
        icon
        :disabled="!dateRange.hasNextPeriod"
        title="Next period"
        @click="dateRange.nextPeriod"
      >
        <v-icon class="small">mdi-chevron-right</v-icon>
      </v-btn>
    </div>

    <v-btn :loading="forceReload.loading" icon @click="dateRange.reload">
      <v-icon>mdi-refresh</v-icon>
    </v-btn>
    <v-btn v-if="!dateRange.isNow" small outlined class="ml-2" @click="dateRange.reloadNow">
      <span>Reset</span>
    </v-btn>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, watchEffect, onMounted, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { injectForceReload } from '@/use/force-reload'

// Components
import PeriodPickerMenu from '@/components/date/PeriodPickerMenu.vue'
import DateRangePickerMenu from '@/components/date/DateRangePickerMenu.vue'

// Misc
import { HOUR } from '@/util/fmt/date'
import { periodsForDays, Period } from '@/models/period'

export default defineComponent({
  name: 'DateRangePicker',
  components: { PeriodPickerMenu, DateRangePickerMenu },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    rangeDays: {
      type: Number,
      default: 3,
    },
    defaultPeriod: {
      type: Number,
      default: HOUR,
    },
  },

  setup(props) {
    const forceReload = injectForceReload()

    const periods = computed(() => {
      return periodsForDays(props.rangeDays)
    })

    onMounted(() => {
      watchEffect(() => {
        if (!periods.value.length || props.dateRange.duration) {
          return
        }

        const period = periods.value.find((p: Period) => p.milliseconds === props.defaultPeriod)
        if (period) {
          props.dateRange.changeDuration(period.milliseconds)
        } else {
          props.dateRange.changeDuration(periods.value[0].milliseconds)
        }
      })
    })

    return { forceReload, periods }
  },
})
</script>

<style lang="scss" scoped></style>
