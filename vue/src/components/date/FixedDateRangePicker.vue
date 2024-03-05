<template>
  <span>
    <span class="text-no-wrap">
      <DateValue :value="around" format="short" class="mr-2 text-subtitle-2" />
      <PeriodPickerMenu
        v-if="dateRange.duration"
        :value="dateRange.duration"
        :periods="periods"
        @input="onInputPeriod"
      />
    </span>
    <v-btn
      v-if="showReload"
      :loading="forceReload.loading"
      small
      outlined
      class="ml-2 px-2"
      @click="dateRange.reload()"
    >
      Reload
    </v-btn>
  </span>
</template>

<script lang="ts">
import { defineComponent, computed, onMounted, watch, watchEffect, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { injectForceReload } from '@/use/force-reload'

// Components
import PeriodPickerMenu from '@/components/date/PeriodPickerMenu.vue'

// Misc
import { periodsForDays } from '@/models/period'
import { HOUR } from '@/util/fmt/date'

export default defineComponent({
  name: 'FixedDateRangePicker',
  components: {
    PeriodPickerMenu,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    rangeDays: {
      type: Number,
      default: 10,
    },
    showReload: {
      type: Boolean,
      default: false,
    },
    around: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const forceReload = injectForceReload()

    const periods = computed(() => {
      return periodsForDays(props.rangeDays)
    })

    onMounted(() => {
      watchEffect(() => {
        if (props.dateRange.duration) {
          return
        }

        let period = periods.value.find((p) => p.milliseconds === HOUR)
        if (!period) {
          period = periods.value[0]
        }
        props.dateRange.changeAround(props.around, period.milliseconds)
      })
    })

    watch(
      () => props.around,
      (date) => {
        props.dateRange.changeAround(date)
      },
      { immediate: true },
    )

    function onInputPeriod(ms: number) {
      props.dateRange.changeAround(props.around, ms)
    }

    return { forceReload, periods, onInputPeriod }
  },
})
</script>

<style lang="scss" scoped></style>
