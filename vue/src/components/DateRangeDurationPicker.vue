<template>
  <v-row no-gutters align="center">
    <v-col>Last</v-col>
    <v-col>
      <v-text-field
        v-model.number="val"
        type="number"
        autofocus
        dense
        flat
        solo
        hide-details
        background-color="grey lighten-4"
        style="width: 78px"
        class="d-inline-block text-body-2 mx-1"
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
        mandatory
        style="width: 116px"
        class="d-inline-block"
        dense
      >
      </v-select>
    </v-col>
    <v-col>
      <v-btn class="primary d-inline-block mx-1" :disabled="!isValid" @click="apply"> Apply </v-btn>
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch } from 'vue'

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
    const val = shallowRef(1)
    const unit = shallowRef(hour)

    const isValid = computed(() => {
      if (val.value * unit.value > 0) {
        return true
      }

      return false
    })

    function apply() {
      emit('input', val.value * unit.value)
    }

    watch(
      () => props.value,
      (value: number) => {
        if (value === val.value * unit.value) {
          return
        }

        for (var i = units.length; i--; ) {
          let ms = units[i].ms
          if (value % ms === 0) {
            val.value = value / ms
            unit.value = ms
            return
          }
        }
      },
      { immediate: true },
    )

    return {
      units,
      val,
      unit,
      isValid,

      apply,
    }
  },
})
</script>

<style lang="scss" scoped></style>
