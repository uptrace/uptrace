<template>
  <v-list dense>
    <v-list-item-group v-model="activePeriodMs" color="primary">
      <v-list-item v-for="item in periods" :key="item.ms" :value="item.ms">
        <v-list-item-content>
          <v-list-item-title>{{ item.text }}</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </v-list-item-group>
  </v-list>
</template>

<script lang="ts">
import {
  defineComponent,
  computed,
  watch,
  watchEffect,
  onMounted,
  PropType,
} from '@vue/composition-api'

// Composables
import { UseDateRange } from '@/use/date-range'

// Utilities
import { Period } from '@/models/period'

export default defineComponent({
  name: 'PeriodList',

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    periods: {
      type: Array as PropType<Period[]>,
      required: true,
    },
  },

  setup(props, { emit }) {
    const activePeriodMs = computed({
      get(): number {
        return props.dateRange.duration
      },
      set(ms: number) {
        props.dateRange.changeDuration(ms)
      },
    })

    const activePeriod = computed((): Period | undefined => {
      const period = props.periods.find((p) => p.ms === activePeriodMs.value)
      return period
    })

    onMounted(() => {
      watchEffect(() => {
        if (activePeriod.value) {
          return
        }

        const period = props.periods[0]
        activePeriodMs.value = period.ms
      })
    })

    watch(
      activePeriod,
      (period) => {
        emit('update:period', period)
      },
      { immediate: true },
    )

    return {
      activePeriodMs,
      activePeriod,
    }
  },
})
</script>
