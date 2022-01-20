<template>
  <XPlaceholder>
    <template v-if="systems.hasNoData" #placeholder>
      <HelpCard :date-range="dateRange" :loading="systems.loading" />
    </template>

    <div>
      <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="pb-0">
        <v-row align="center" justify="space-between" class="mb-4">
          <v-col cols="auto">
            <SystemPicker :date-range="dateRange" :systems="systems" route-name="GroupList" />
          </v-col>

          <v-col cols="auto">
            <DateRangePicker :date-range="dateRange" />
          </v-col>
        </v-row>
      </v-container>
    </div>

    <v-container :fluid="$vuetify.breakpoint.mdAndDown" class="pt-2">
      <UptraceQuery :uql="uql" class="mb-1">
        <SpanFilters :uql="uql" :systems="systems" :axios-params="axiosParams" />
      </UptraceQuery>

      <GroupList
        :date-range="dateRange"
        :systems="systems"
        :uql="uql"
        :axios-params="axiosParams"
      />
    </v-container>
  </XPlaceholder>
</template>

<script lang="ts">
import { defineComponent, computed } from '@vue/composition-api'

// Composables
import { useTitle } from '@vueuse/core'
import { useDateRange } from '@/use/date-range'
import { useSystems } from '@/use/systems'
import { useUql, buildGroupBy } from '@/use/uql'

// Components
import DateRangePicker from '@/components/DateRangePicker.vue'
import SystemPicker from '@/components/SystemPicker.vue'
import HelpCard from '@/components/HelpCard.vue'
import UptraceQuery from '@/components/uql/UptraceQuery.vue'
import SpanFilters from '@/components/uql/SpanFilters.vue'
import GroupList from '@/components/GroupList.vue'

// Utilities
import { xkey } from '@/models/otelattr'

export default defineComponent({
  name: 'Tracing',
  components: {
    DateRangePicker,
    SystemPicker,
    HelpCard,
    UptraceQuery,
    SpanFilters,
    GroupList,
  },

  setup() {
    useTitle('Explore spans')
    const dateRange = useDateRange()
    const systems = useSystems(dateRange)
    const uql = useUql({
      query: buildGroupBy(xkey.spanGroupId),
      syncQuery: true,
    })

    const axiosParams = computed(() => {
      return {
        ...dateRange.axiosParams(),
        ...uql.axiosParams(),
        system: systems.activeValue,
      }
    })

    return { dateRange, systems, uql, axiosParams }
  },
})
</script>

<style lang="scss" scoped></style>
