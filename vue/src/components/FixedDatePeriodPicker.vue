<template>
  <span>
    <span class="text-no-wrap">
      <XDate :date="date" format="short" class="mr-2 text-subtitle-2" />
      <PeriodPickerMenu :value="dateRange.duration" :periods="periods" @input="onInputPeriod" />
    </span>
    <v-btn v-if="withReload" small outlined class="ml-2" @click="dateRange.forceReload">
      <v-icon small left>mdi-refresh</v-icon>
      <span>Reload</span>
    </v-btn>
  </span>
</template>

<script lang="ts">
import { defineComponent, computed, onMounted, watchEffect, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import PeriodPickerMenu from '@/components/date/PeriodPickerMenu.vue'

// Utilities
import { periodsForDays } from '@/models/period'
import { hour } from '@/util/fmt/date'

export default defineComponent({
  name: 'FixedDatePeriodPicker',
  components: {
    PeriodPickerMenu,
  },

  props: {
    date: {
      type: String,
      required: true,
    },
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    rangeDays: {
      type: Number,
      default: 10,
    },
    withReload: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const periods = computed(() => {
      return periodsForDays(props.rangeDays)
    })

    function onInputPeriod(ms: number) {
      props.dateRange.changeWithin(new Date(props.date), ms)
    }

    onMounted(() => {
      watchEffect(() => {
        if (props.dateRange.duration) {
          return
        }

        const period = periods.value.find((p) => p.ms === hour)
        if (period) {
          props.dateRange.changeDuration(period.ms)
          return
        }

        props.dateRange.changeDuration(periods.value[0].ms)
      })
    })

    return { periods, onInputPeriod }
  },
})
</script>

<style lang="scss" scoped></style>
