<template>
  <div>
    <template v-if="systems.dataHint?.before || systems.dataHint?.after">
      <PageToolbar>
        <v-spacer />
        <DateRangePicker :date-range="dateRange" :range-days="30" />
      </PageToolbar>

      <v-card flat class="text-center" style="margin-top: 180px">
        <v-icon v-show="!systems.loading" size="48">mdi-calendar-range</v-icon>
        <v-progress-circular v-show="systems.loading" color="purple" size="48" indeterminate />

        <v-row align="center" justify="center" no-gutters>
          <v-col cols="auto">
            <v-btn
              :disabled="!systems.dataHint?.before"
              x-large
              icon
              title="Previous period with data"
              @click="changeAround(systems.dataHint?.before)"
              ><v-icon size="48">mdi-chevron-left</v-icon></v-btn
            >
          </v-col>
          <v-col cols="auto">
            <v-card width="360" flat>
              <v-card-text>
                There are no results for the selected date range.<br />
                Use <strong>arrows</strong> to jump to the periods with data.
              </v-card-text>
            </v-card>
          </v-col>
          <v-col cols="auto">
            <v-btn
              :disabled="!systems.dataHint?.after"
              icon
              x-large
              title="Next period with data"
              @click="changeAround(systems.dataHint?.after)"
              ><v-icon size="48">mdi-chevron-right</v-icon></v-btn
            >
          </v-col>
        </v-row>
      </v-card>
    </template>

    <HelpCard v-else-if="systems.status.hasData()" :loading="systems.loading" show-reload />

    <v-skeleton-loader v-else type="card,table"></v-skeleton-loader>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useRoute, useSyncQueryParams } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseSystems } from '@/tracing/system/use-systems'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import HelpCard from '@/tracing/HelpCard.vue'

export default defineComponent({
  name: 'TracingPlaceholder',
  components: { DateRangePicker, HelpCard },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
  },

  setup(props) {
    const route = useRoute()

    useSyncQueryParams({
      fromQuery(queryParams) {
        props.dateRange.parseQueryParams(queryParams)
      },
      toQuery() {
        return {
          ...route.value.query,
          ...props.dateRange.queryParams(),
        }
      },
    })

    function changeAround(dt: string) {
      props.dateRange.changeAround(dt)
    }

    return { changeAround }
  },
})
</script>

<style lang="scss" scoped></style>
