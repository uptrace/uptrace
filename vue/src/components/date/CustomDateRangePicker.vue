<template>
  <div>
    <v-row no-gutters>
      <v-col class="pt-3 text-body-1">From date</v-col>
      <v-col cols="auto">
        <DateTextInput v-model="gte" @input="updateLt" />
      </v-col>
    </v-row>
    <v-row no-gutters>
      <v-col class="pt-3 text-body-1">To date</v-col>
      <v-col cols="auto">
        <DateTextInput v-model="lt" />
      </v-col>
    </v-row>
    <v-row>
      <v-col></v-col>
      <v-col cols="auto">
        <v-spacer></v-spacer>
        <v-btn class="primary" :disabled="!isValid" @click="apply">Apply</v-btn>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { addMilliseconds } from 'date-fns'
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import DateTextInput from '@/components/date/DateTextInput.vue'

// Misc
import { HOUR } from '@/util/fmt/date'

export default defineComponent({
  name: 'CustomDateRangePicker',
  components: { DateTextInput },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
  },

  setup(props, { emit }) {
    const gte = shallowRef(new Date(Date.now() - HOUR))
    const lt = shallowRef(new Date())

    const isValid = computed((): boolean => {
      return gte.value! < lt.value!
    })

    watch(
      () => props.dateRange.gte,
      (date: Date | undefined) => {
        if (date) {
          gte.value = date
        }
      },
      { immediate: true },
    )

    watch(
      () => props.dateRange.lt,
      (date: Date | undefined) => {
        if (date) {
          lt.value = date
        }
      },
      { immediate: true },
    )

    function updateLt(gteValue: Date) {
      lt.value = addMilliseconds(gteValue, props.dateRange.duration)
    }

    function apply() {
      emit('input', gte.value, lt.value)
    }

    return {
      gte,
      lt,
      isValid,

      apply,
      updateLt,
    }
  },
})
</script>

<style lang="scss" scoped></style>
