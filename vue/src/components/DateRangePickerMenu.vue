<template>
  <v-menu v-model="menu" offset-y transition="slide-x-transition" :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <span class="mr-2">
        <v-btn icon v-bind="attrs" v-on="on">
          <v-icon small>mdi-calendar-blank</v-icon>
        </v-btn>
        <v-btn v-if="dateRange.hasNextPeriod" text small class="px-1" v-bind="attrs" v-on="on">
          <span><XDate :date="dateRange.gte" :format="format" /> - </span>
          <XDate :date="dateRange.lt" :format="format" />
        </v-btn>
      </span>
    </template>
    <v-card width="auto">
      <div @click="byDuration = true">
        <v-card :disabled="!byDuration" tile outlined color="transparent">
          <v-card-text>
            <DateRangeDurationPicker v-model="duration" />
          </v-card-text>
        </v-card>
      </div>
      <v-divider></v-divider>
      <div @click="byDuration = false">
        <v-card :disabled="byDuration" tile outlined color="transparent">
          <v-card-text>
            <v-row no-gutters>
              <v-col class="pt-3 pr-2">From</v-col>
              <v-col cols="auto" class="d-flex justify-end">
                <DateTextInput v-model="gte" />
              </v-col>
            </v-row>
            <v-row no-gutters>
              <v-col class="pt-3 pr-2">To</v-col>
              <v-col cols="auto" class="d-flex justify-end">
                <DateTextInput v-model="lt" />
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
      </div>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn class="primary mr-1" :disabled="!isValid" @click="apply">Apply</v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, ref, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import DateTextInput from '@/components/DateTextInput.vue'
import DateRangeDurationPicker from '@/components/DateRangeDurationPicker.vue'

// Utilities
import { hour } from '@/util/date'

export default defineComponent({
  name: 'DateRangePickerMenu',
  components: { DateTextInput, DateRangeDurationPicker },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    format: {
      type: String,
      default: 'short',
    },
  },

  setup(props) {
    const menu = ref(false)
    const gte = shallowRef(new Date(Date.now() - hour))
    const lt = shallowRef(new Date())
    const byDuration = shallowRef(true)
    const duration = shallowRef(hour)

    const isValid = computed((): boolean => {
      if (byDuration.value && duration.value !== 0) {
        true
      }

      return gte.value! < lt.value!
    })

    function apply() {
      if (byDuration.value) {
        props.dateRange.changeDuration(duration.value)
        return
      }

      props.dateRange.change(gte.value, lt.value)
      menu.value = false
    }

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

    watch(
      () => props.dateRange.duration,
      (ms) => {
        if (ms) {
          duration.value = ms
        }
      },
      { immediate: true },
    )

    return {
      menu,
      byDuration,
      duration,
      gte,
      lt,
      isValid,

      apply,
    }
  },
})
</script>

<style lang="scss" scoped></style>
