<template>
  <v-menu v-model="menu" offset-y transition="slide-x-transition" :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-tooltip bottom>
        <template #activator="tooltip">
          <v-btn
            v-if="dateRange.isNow && !forceDateRange"
            icon
            v-bind="{ ...tooltip.attrs, ...attrs }"
            v-on="{ ...tooltip.on, ...on }"
          >
            <v-icon small>mdi-calendar-blank</v-icon>
          </v-btn>
          <v-btn v-else-if="dateRange.isValid" text small class="px-2" v-bind="attrs" v-on="on">
            <v-icon small left>mdi-calendar-blank</v-icon>
            <DateRange :start="dateRange.gte" :end="dateRange.lt" />
          </v-btn>
        </template>
        <DateRange :start="dateRange.gte" :end="dateRange.lt" />
      </v-tooltip>
    </template>

    <v-card width="auto">
      <v-card-text class="pa-5">
        <CustomDurationPicker :value="dateRange.duration" @input="applyDuration" />

        <div class="my-5 d-flex align-center">
          <v-divider />
          <div class="mx-2 text-subtitle-1 grey--text text--lighten-1">or</div>
          <v-divider />
        </div>

        <CustomDateRangePicker :date-range="dateRange" @input="applyRange" />
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
import DateRange from '@/components/date/DateRange.vue'

export default defineComponent({
  name: 'DateRangePickerMenu',
  components: {
    CustomDurationPicker,
    CustomDateRangePicker,
    DateRange,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    forceDateRange: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const menu = ref(false)

    function applyDuration(ms: number) {
      props.dateRange.changeDuration(ms)
      props.dateRange.reloadNow()
      menu.value = false
    }

    function applyRange(gte: Date, lt: Date) {
      props.dateRange.change(gte, lt)
      menu.value = false
    }

    return {
      menu,

      applyDuration,
      applyRange,
    }
  },
})
</script>

<style lang="scss" scoped></style>
