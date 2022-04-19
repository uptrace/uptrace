<template>
  <div>
    <v-card outlined rounded="lg" class="mb-4">
      <v-toolbar flat color="light-blue lighten-5">
        <v-toolbar-title>Hosts</v-toolbar-title>
        <v-toolbar-items class="ml-5">
          <v-col align-self="center">
            <SystemPicker
              :date-range="dateRange"
              :systems="systems"
              route-name="HostOverview"
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
          :loading="hosts.loading"
          :items="hosts.pageHosts"
          :order="hosts.order"
          column="host"
          :attribute="xkey.hostName"
          :base-column-route="groupListRoute"
        />
      </v-card-text>
    </v-card>

    <XPagination :pager="hosts.pager" />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from '@vue/composition-api'

// Composables
import { UseSystems } from '@/use/systems'
import { useHosts } from '@/use/hosts'
import { UseDateRange } from '@/use/date-range'
import { buildGroupBy } from '@/use/uql'

// Components
import OverviewTable from '@/components/OverviewTable.vue'
import SystemPicker from '@/components/SystemPicker.vue'

// Utilities
import { xkey } from '@/models/otelattr'

export default defineComponent({
  name: 'HostOverview',
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
    const hosts = useHosts(props.dateRange, props.systems)

    const query = computed(() => {
      return buildGroupBy(xkey.hostName)
    })

    const groupListRoute = computed(() => {
      return {
        name: 'GroupList',
        query: {
          ...props.dateRange.queryParams(),
          ...props.systems.axiosParams(),
          query: query.value,
        },
      }
    })

    return {
      xkey,

      hosts,
      groupListRoute,
    }
  },
})
</script>

<style lang="scss"></style>
