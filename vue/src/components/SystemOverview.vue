<template>
  <XPlaceholder>
    <v-card outlined rounded="lg" class="mb-4">
      <v-toolbar flat color="light-blue lighten-5">
        <v-toolbar-title>Systems</v-toolbar-title>
        <v-spacer />
        <v-btn :to="exploreRoute" small class="primary">Open explorer</v-btn>
      </v-toolbar>

      <v-card-text class="pb-0">
        <v-slide-group
          v-if="systems.types.length >= 3"
          v-model="systems.filter"
          center-active
          show-arrows
          class="ml-2"
        >
          <v-slide-item
            v-for="(type, i) in systems.types"
            :key="type"
            v-slot="{ active, toggle }"
            :value="type"
          >
            <v-btn
              :input-value="active"
              active-class="blue white--text"
              small
              depressed
              rounded
              :class="{ 'ml-1': i > 0 }"
              @click="toggle"
            >
              {{ type }}
            </v-btn>
          </v-slide-item>
        </v-slide-group>
      </v-card-text>

      <v-card-text>
        <OverviewTable
          :date-range="dateRange"
          :loading="systems.loading"
          :items="systems.pageSystems"
          :order="systems.order"
        >
        </OverviewTable>
      </v-card-text>
    </v-card>

    <XPagination :pager="systems.pager" />
  </XPlaceholder>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from '@vue/composition-api'

// Composables
import { UseDateRange } from '@/use/date-range'
import { buildGroupBy } from '@/use/uql'
import { useSystemStats } from '@/use/system-stats'

// Components
import OverviewTable from '@/components/OverviewTable.vue'

// Utilities
import { xkey } from '@/models/otelattr'

export default defineComponent({
  name: 'SystemOverview',
  components: { OverviewTable },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
  },

  setup(props) {
    const systems = useSystemStats(props.dateRange)

    const exploreRoute = computed(() => {
      return {
        name: 'GroupList',
        query: {
          system: xkey.allSystem,
          query: buildGroupBy(xkey.spanSystem),
        },
      }
    })

    return {
      xkey,

      systems,
      exploreRoute,
    }
  },
})
</script>

<style lang="scss">
.v-data-table.large > .v-data-table__wrapper > table {
  & > tbody > tr > td {
    height: 60px;
  }
}
</style>
