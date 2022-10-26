<template>
  <v-menu v-model="menu" offset-y transition="slide-x-transition" :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn v-if="dateRange.isNow" icon v-bind="attrs" v-on="on">
        <v-icon small>mdi-calendar-blank</v-icon>
      </v-btn>
      <v-btn v-else-if="dateRange.isValid" text small class="px-1" v-bind="attrs" v-on="on">
        <v-icon small left>mdi-calendar-blank</v-icon>
        <span><XDate :date="dateRange.gte" :format="format" /> - </span>
        <XDate :date="dateRange.lt" :format="format" />
      </v-btn>
    </template>

    <v-card width="auto">
      <v-card-text class="pa-5">
        <CustomDurationPicker :value="dateRange.duration" @input="applyDuration" />

        <div class="my-5 d-flex align-center">
          <v-divider />
          <div class="mx-2 grey--text text--lighten-1">or</div>
          <v-divider />
        </div>

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
import CustomDurationPicker from '@/components/date/CustomDurationPicker.vue'
import CustomDateRangePicker from '@/components/date/CustomDateRangePicker.vue'

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
