<template>
  <div class="text-no-wrap">
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

    <v-btn small outlined class="ml-2" @click="dateRange.reload">
      <v-icon small left>mdi-refresh</v-icon>
      <span>Reload</span>
    </v-btn>
    <v-btn v-if="!dateRange.isNow" small outlined class="ml-2" @click="dateRange.reloadNow">
      <span>Now</span>
    </v-btn>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, watchEffect, onMounted, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import PeriodPickerMenu from '@/components/date/PeriodPickerMenu.vue'
import DateRangePickerMenu from '@/components/date/DateRangePickerMenu.vue'

// Utilities
import { HOUR } from '@/util/fmt/date'
import { periodsForDays } from '@/models/period'

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
    const periods = computed(() => {
      return periodsForDays(props.rangeDays)
    })

    onMounted(() => {
      watchEffect(() => {
        if (props.dateRange.duration) {
          return
        }

        const period = periods.value.find((p) => p.milliseconds === props.defaultPeriod)
        if (period) {
          props.dateRange.changeDuration(period.milliseconds)
          return
        }

        props.dateRange.changeDuration(periods.value[0].milliseconds)
      })
    })

    return { periods }
  },
})
</script>

<style lang="scss" scoped></style>
