<template>
  <v-menu offset-y>
    <template #activator="{ on }">
      <v-btn text small class="px-1" v-on="on">
        <span v-if="activePeriod">{{ activePeriod.text }}</span>
        <span v-else>Period</span>
        <v-icon>mdi-menu-down</v-icon>
      </v-btn>
    </template>

    <v-list dense>
      <v-list-item-group :value="value" @change="onChange">
        <v-list-item v-for="item in periods" :key="item.ms" :value="item.ms">
          <v-list-item-content>
            <v-list-item-title>Last {{ item.text }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list-item-group>
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from '@vue/composition-api'

// Utilities
import { Period } from '@/models/period'
import { formatDistance } from 'date-fns'

export default defineComponent({
  name: 'PeriodPickerMenu',

  props: {
    value: {
      type: Number,
      required: true,
    },
    periods: {
      type: Array as PropType<Period[]>,
      required: true,
    },
  },

  setup(props, { emit }) {
    const activePeriod = computed((): Period | undefined => {
      const period = props.periods.find((p) => p.ms === props.value)

      if (!period) {
        return {
          text: formatDistance(0, props.value),
          ms: props.value,
        }
      }

      return period
    })

    function onChange(ms: number) {
      emit('input', ms)
    }

    return { activePeriod, onChange }
  },
})
</script>

<style lang="scss" scoped></style>
