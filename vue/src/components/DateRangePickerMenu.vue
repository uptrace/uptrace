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
      <v-card-text class="pa-5">
        <CustomDurationPicker :value="dateRange.duration" @input="applyDuration" />

        <v-divider class="my-6"></v-divider>

        <CustomDateRangePicker :date-range="dateRange" @input="applyPeriod" />
      </v-card-text>
    </v-card>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, ref, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import CustomDurationPicker from '@/components/CustomDurationPicker.vue'
import CustomDateRangePicker from '@/components/CustomDateRangePicker.vue'

export default defineComponent({
  name: 'DateRangePickerMenu',
  components: {
    CustomDurationPicker,
    CustomDateRangePicker,
  },

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

    function applyDuration(ms: number) {
      props.dateRange.changeDuration(ms)
      menu.value = false
    }

    function applyPeriod(gte: Date, lt: Date) {
      props.dateRange.change(gte, lt)
      menu.value = false
    }

    return {
      menu,

      applyDuration,
      applyPeriod,
    }
  },
})
</script>

<style lang="scss" scoped></style>
