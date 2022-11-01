<template>
  <div>
    <v-card outlined rounded="lg" class="mb-4">
      <v-toolbar flat color="light-blue lighten-5">
        <v-toolbar-title>{{ attr }} overview</v-toolbar-title>
        <v-toolbar-items class="ml-5">
          <v-col align-self="center">
            <SystemPicker
              v-model="systems.activeSystem"
              :loading="systems.loading"
              :items="systems.items"
              :all-system="SystemName.all"
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
          :loading="overview.loading"
          :items="overview.pagedGroups"
          :order="overview.order"
          :attr="attr"
          :base-item-route="spanListRoute"
        />
      </v-card-text>
    </v-card>

    <XPagination :pager="overview.pager" />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseEnvs, UseServices } from '@/tracing/use-quick-span-filters'
import { UseSystems } from '@/tracing/system/use-systems'
import { createUqlEditor } from '@/use/uql'
import { useOverview } from '@/tracing/use-overview'

// Components
import SystemPicker from '@/tracing/system/SystemPicker.vue'
import OverviewTable from '@/tracing/overview/OverviewTable.vue'

// Utilities
import { SystemName } from '@/models/otel'

export default defineComponent({
  name: 'AttrOverview',
  components: { SystemPicker, OverviewTable },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    envs: {
      type: Object as PropType<UseEnvs>,
      required: true,
    },
    services: {
      type: Object as PropType<UseServices>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()

    const attr = computed(() => {
      return route.value.params.attr
    })

    const overview = useOverview(() => {
      return {
        ...props.dateRange.axiosParams(),
        ...props.envs.axiosParams(),
        ...props.services.axiosParams(),
        ...props.systems.axiosParams(),
        attr: attr.value,
      }
    })

    const query = computed(() => {
      return createUqlEditor().exploreAttr(attr.value).toString()
    })

    const groupListRoute = computed(() => {
      return {
        name: 'SpanGroupList',
        query: {
          ...route.value.query,
          ...props.systems.queryParams(),
          query: query.value,
        },
      }
    })

    const spanListRoute = computed(() => {
      return {
        name: 'SpanList',
        query: {
          ...route.value.query,
          ...props.systems.queryParams(),
          query: query.value,
        },
      }
    })

    return {
      SystemName,
      attr,
      overview,
      groupListRoute,
      spanListRoute,
    }
  },
})
</script>

<style lang="scss"></style>
