<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn v-if="dateRange.isNow" icon v-bind="attrs" v-on="on">
        <v-icon small :title="datetimeFull(dateRange.gte)">mdi-calendar-blank</v-icon>
      </v-btn>
      <v-btn v-else text small class="px-1" v-bind="attrs" v-on="on">
        <XDate :date="dateRange.gte" :format="format" />
        <v-icon>mdi-menu-down</v-icon>
      </v-btn>
    </template>

    <DateTimePeriod :date-range="dateRange" :periods="periods" @click:ok="menu = false" />
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, ref, PropType } from '@vue/composition-api'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import DateTimePeriod from '@/components/DateTimePeriod.vue'

// Utilities
import { datetimeFull } from '@/util/date'
import { Period } from '@/models/period'

export default defineComponent({
  name: 'DateTimePeriodMenu',
  components: { DateTimePeriod },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    periods: {
      type: Array as PropType<Period[]>,
      required: true,
    },
    format: {
      type: String,
      default: 'short',
    },
  },

  setup() {
    const menu = ref(false)

    return {
      menu,

      datetimeFull,
    }
  },
})
</script>

<style lang="scss" scoped></style>
