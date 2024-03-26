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

    <v-btn
      :loading="forceReload.loading"
      small
      outlined
      class="ml-1 px-2"
      @click="dateRange.reload"
    >
      Reload
    </v-btn>
    <v-btn
      v-if="dateRange.isNow"
      icon
      :title="dateRange.autoReloadEnabled ? 'Stop auto-reloading' : 'Start auto-reloading'"
      class="ml-1"
      @click="dateRange.toggleAutoReload()"
    >
      <v-icon>{{ dateRange.autoReloadEnabled ? 'mdi-pause' : 'mdi-play' }}</v-icon>
    </v-btn>
    <v-btn v-else small outlined class="ml-2 px-2" @click="dateRange.reloadNow"> Reset </v-btn>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, watchEffect, onMounted, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { injectForceReload } from '@/use/force-reload'

// Components
import DateRangePickerMenu from '@/components/date/DateRangePickerMenu.vue'
import PeriodPickerMenu from '@/components/date/PeriodPickerMenu.vue'

// Misc
import { periodsForDays, Period } from '@/models/period'
import { HOUR } from '@/util/fmt'

export default defineComponent({
  name: 'DateRangePicker',
  components: { DateRangePickerMenu, PeriodPickerMenu },

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
