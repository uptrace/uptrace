<template>
  <div>
    Last
    <v-text-field
      v-model.number="val"
      type="number"
      autofocus
      dense
      flat
      solo
      hide-details
      background-color="grey lighten-4"
      style="width: 85px"
      class="d-inline-block text-body-2 mx-2"
    >
    </v-text-field>
    <v-select
      v-model="unit"
      :items="units"
      item-text="name"
      item-value="ms"
      hide-details
      outlined
      mandatory
      style="width: 140px"
      class="d-inline-block"
      dense
    >
    </v-select>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch } from 'vue'

// Utilities
import { minute, hour, day } from '@/util/date'

export default defineComponent({
  name: 'DateRangeDurationPicker',

  props: {
    value: {
      type: Number,
      required: true,
    },
  },

  setup(props, { emit }) {
    const units = [
      {
        name: 'minutes',
        ms: minute,
        maxVal: 60,
      },
      {
        name: 'hours',
        ms: hour,
        maxVal: 24,
      },
      {
        name: 'days',
        ms: day,
        maxVal: 31,
      },
      {
        name: 'week',
        ms: 7 * day,
        maxVal: 10000,
      },
    ]

    const val = shallowRef(0)
    const unit = shallowRef(0)

    watch(
      () => props.value,
      (ms: number) => {
        if (!ms) {
          return
        }

        if (!unit.value) {
          for (let u of units) {
            if (ms / u.ms < u.maxVal) {
              unit.value = u.ms
              val.value = ms / u.ms
              return
            }
          }
        }
      },
      { immediate: true },
    )

    watch(
      () => val.value * unit.value,
      (ms, oldMs: number) => {
        if (ms !== oldMs) {
          emit('input', ms)
        }
      },
    )

    return {
      units,
      val,
      unit,
    }
  },
})
</script>

<style lang="scss" scoped></style>
