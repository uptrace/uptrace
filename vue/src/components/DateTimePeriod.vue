<template>
  <v-card width="340">
    <v-tabs v-model="activeTab" background-color="primary" dark icons-and-text fixed-tabs>
      <v-tab href="#date">
        Date
        <v-icon>mdi-calendar-outline</v-icon>
      </v-tab>
      <v-tab href="#time">
        Time
        <v-icon>mdi-clock-outline</v-icon>
      </v-tab>
      <v-tab href="#period">
        Period
        <v-icon>mdi-calendar-range-outline</v-icon>
      </v-tab>
    </v-tabs>

    <v-tabs-items v-model="activeTab">
      <v-tab-item value="date">
        <v-date-picker v-model="dateRange.datePicker" full-width></v-date-picker>
      </v-tab-item>
      <v-tab-item value="time">
        <v-time-picker
          v-model="dateRange.timePicker"
          format="24hr"
          full-width
          class="v-time-picker-custom"
        ></v-time-picker>
      </v-tab-item>
      <v-tab-item value="period">
        <v-card min-height="378px">
          <PeriodList :date-range="dateRange" :periods="periods" />
        </v-card>
      </v-tab-item>
    </v-tabs-items>

    <v-card-actions>
      <v-spacer />
      <v-btn text color="primary" @click="$emit('click:ok')">OK</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, ref, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import PeriodList from '@/components/PeriodList.vue'

// Utilities
import { Period } from '@/models/period'

export default defineComponent({
  name: 'DateTimePeriod',
  components: { PeriodList },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    periods: {
      type: Array as PropType<Period[]>,
      required: true,
    },
  },

  setup() {
    const activeTab = ref()
    return { activeTab }
  },
})
</script>

<style lang="scss">
.v-time-picker-custom {
  .v-picker__title {
    height: 88px;
    padding-top: 10px;
  }
}
</style>
