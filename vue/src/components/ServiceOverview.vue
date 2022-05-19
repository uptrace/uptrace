<template>
  <div>
    <v-card outlined rounded="lg" class="mb-4">
      <v-toolbar flat color="light-blue lighten-5">
        <v-toolbar-title>Services</v-toolbar-title>
        <v-toolbar-items class="ml-5">
          <v-col align-self="center">
            <SystemPicker
              :date-range="dateRange"
              :systems="systems"
              :tree="systems.tree"
              route-name="ServiceOverview"
              outlined
            />
          </v-col>
        </v-toolbar-items>

        <v-spacer />
        <v-btn :to="groupListRoute" small class="primary">View in explorer</v-btn>
      </v-toolbar>

      <v-card-text>
        <OverviewTable
          :date-range="dateRange"
          :loading="services.loading"
          :items="services.pageServices"
          :order="services.order"
          column="service"
          :attribute="xkey.serviceName"
          :base-column-route="groupListRoute"
        />
      </v-card-text>
    </v-card>

    <XPagination :pager="services.pager" />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from '@vue/composition-api'

// Composables
import { UseSystems } from '@/use/systems'
import { useServices } from '@/use/services'
import { UseDateRange } from '@/use/date-range'
import { buildGroupBy } from '@/use/uql'

// Components
import OverviewTable from '@/components/OverviewTable.vue'
import SystemPicker from '@/components/SystemPicker.vue'

// Utilities
import { xkey } from '@/models/otelattr'

export default defineComponent({
  name: 'ServiceOverview',
  components: { OverviewTable, SystemPicker },

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
    const services = useServices(props.dateRange, props.systems)

    const query = computed(() => {
      return buildGroupBy(xkey.serviceName)
    })

    const groupListRoute = computed(() => {
      return {
        name: 'SpanGroupList',
        query: {
          ...props.dateRange.queryParams(),
          ...props.systems.axiosParams(),
          query: query.value,
        },
      }
    })

    return {
      xkey,

      services,
      groupListRoute,
    }
  },
})
</script>

<style lang="scss"></style>
