<template>
  <v-menu offset-y>
    <template #activator="{ on }">
      <v-btn text small class="px-1" v-on="on">
        <XDuration v-if="!activePeriod.byDateRange" :duration="activePeriod.ms" fixed />
        <span v-else>
          <XDate :date="activePeriod.from" format="short" /> -
          <XDate :date="activePeriod.to" format="short" />
        </span>
        <v-icon>mdi-menu-down</v-icon>
      </v-btn>
    </template>

    <v-card dense @click.stop>
      <v-card-text @click="activePeriod.byDateRange = false">
        Last
        <v-text-field
          v-model.number="unitVal"
          type="number"
          autofocus
          dense
          flat
          solo
          hide-details="auto"
          background-color="grey lighten-4"
          :disabled="activePeriod.byDateRange"
          :rules="form.numberRules"
          style="width: 70px"
          class="d-inline-block text-body-2 mr-1"
        >
        </v-text-field>
        <v-btn-toggle v-model="unitIndex" :disabled="activePeriod.byDateRange" dense mandatory>
          <v-btn v-for="unit in units" :key="unit.name">
            {{ unit.name }}
          </v-btn>
        </v-btn-toggle>
      </v-card-text>

      <v-card-text @click="activePeriod.byDateRange = true">
        <p>Custom time range</p>
        From <DateRangePickerInput v-model="activePeriod.from" /> To
        <DateRangePickerInput v-model="activePeriod.to" />
      </v-card-text>
      <v-card-actions>
        <v-btn @click="apply"> Save </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, proxyRefs, watch, PropType } from '@vue/composition-api'

// Composables
import { UseDateRange } from '@/use/date-range'

// Utilities
import { minute, hour, day } from '@/util/date'
import DateRangePickerInput from '@/components/DateRangePickerInput.vue'

export interface Unit {
  name: string
  ms: number
}

export interface Period {
  from: Date
  to: Date
  ms: number
  byDateRange: boolean
}

export default defineComponent({
  name: 'TimePeriodPickerMenu',
  components: { DateRangePickerInput },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
  },

  setup(props) {
    const units = [
      {
        name: 'minutes',
        ms: minute,
      },
      {
        name: 'hours',
        ms: hour,
      },
      {
        name: 'days',
        ms: day,
      },
      {
        name: 'week',
        ms: 7 * day,
      },
    ]
    const unitIndex = shallowRef(1)
    const unitVal = shallowRef(1)

    const activePeriod = shallowRef<Period>({
      from: new Date(Date.now() - 15 * minute),
      to: new Date(Date.now()),
      ms: 15 * minute,
      byDateRange: false,
    })

    function apply() {
      if (!activePeriod.value.byDateRange) {
        props.dateRange.changeDuration(activePeriod.value.ms)
        return
      }

      props.dateRange.change(activePeriod.value.from, activePeriod.value.to)
    }

    watch(
      () => unitVal.value * units[unitIndex.value].ms,
      (ms: number) => {
        activePeriod.value.byDateRange = false
        activePeriod.value.ms = ms
      },
    )

    return {
      activePeriod,
      unitIndex,
      unitVal,
      units,
      form: useForm(),

      apply,
    }
  },
})

function useForm() {
  const isValid = shallowRef(false)

  const numberRules = [
    (v: any) => {
      const n = parseFloat(v)
      if (isNaN(n)) {
        return 'must be a number'
      }
      return true
    },
  ]

  return proxyRefs({ isValid, numberRules })
}
</script>

<style lang="scss" scoped></style>
