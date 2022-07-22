<template>
  <v-row dense align="center">
    <v-col class="text-body-1">Last</v-col>
    <v-col>
      <v-text-field
        v-model.number="amount"
        type="number"
        hide-details
        outlined
        dense
        mandatory
        style="width: 78px"
      >
      </v-text-field>
    </v-col>
    <v-col>
      <v-select
        v-model="unit"
        :items="units"
        item-text="name"
        item-value="ms"
        hide-details
        outlined
        dense
        mandatory
        style="width: 118px"
      >
      </v-select>
    </v-col>
    <v-col>
      <v-btn class="primary d-inline-block mx-1" :disabled="!isValid" @click="apply">Apply</v-btn>
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch } from 'vue'

// Utilities
import { minute, hour, day } from '@/util/date'

interface Unit {
  name: string
  ms: number
}

export default defineComponent({
  name: 'DateRangeDurationPicker',

  props: {
    value: {
      type: Number,
      required: true,
    },
  },

  setup(props, { emit }) {
    const units: Unit[] = [
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
    ]

    const amount = shallowRef(1)
    const unit = shallowRef(hour)

    const isValid = computed(() => {
      if (amount.value * unit.value > 0) {
        return true
      }

      return false
    })

    function apply() {
      emit('input', amount.value * unit.value)
    }

    watch(
      () => props.value,
      (ms: number) => {
        const found = findUnit(ms)
        amount.value = Math.floor(ms / found.ms)
        unit.value = found.ms
      },
      { immediate: true },
    )

    function findUnit(ms: number): Unit {
      for (let i = units.length - 1; i >= 0; i--) {
        const unit = units[i]

        if (ms / unit.ms >= 1000) {
          const found = units[i + 1]
          if (found) {
            return found
          }
        }

        if (ms % unit.ms === 0) {
          return unit
        }
      }

      return units[0]
    }

    return {
      units,
      amount,
      unit,
      isValid,

      apply,
    }
  },
})
</script>

<style lang="scss" scoped></style>
