<template>
  <div>
    <v-row v-if="dashboard.entries.length">
      <v-col>
        <v-card outlined rounded="lg" class="py-2 px-4">
          <DashQueryBuilder :date-range="dateRange" :metrics="metricNames" :uql="uql" class="mb-1">
            <template #prepend-actions>
              <v-btn
                v-if="tableQueryMan.isDirty"
                :loading="tableQueryMan.pending"
                small
                depressed
                class="mr-4"
                @click="tableQueryMan.save"
              >
                <v-icon small left>mdi-check</v-icon>
                <span>Save</span>
              </v-btn>
            </template>
          </DashQueryBuilder>
        </v-card>
      </v-col>
    </v-row>

    <v-row v-if="dashboard.gridGauges.length" :dense="$vuetify.breakpoint.mdAndDown" class="mt-4">
      <v-col v-for="gauge in dashboard.gridGauges" :key="gauge.id" cols="auto">
        <DashGaugeCard :date-range="dateRange" :gauge="gauge" :base-query="baseQuery" />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <draggable
          v-model="dashboard.entries"
          handle=".draggable-handle"
          class="row row--dense"
          @end="dashEntryMan.updateOrder(dashboard.entries)"
        >
          <DashGridEntry
            v-for="(entry, i) in dashboard.entries"
            :key="entry.id"
            :date-range="dateRange"
            :metrics="metrics"
            :dashboard="dashboard"
            :dash-entry="entry"
            :base-query="baseQuery"
            :editing="entry.id === ''"
            @input:base-query="onBaseQuery(i, $event)"
            @change="dashboard.reload()"
          />

          <template v-if="!dashboard.isTemplate && !dashboard.isFull" #footer>
            <v-col cols="12" md="6">
              <v-card
                height="100%"
                min-height="260"
                outlined
                rounded="lg"
                class="d-flex align-center"
              >
                <v-card-text class="text-center">
                  <v-icon size="90" color="grey lighten-2" @click="dashboard.addGridEntry"
                    >mdi-plus</v-icon
                  >
                </v-card-text>
              </v-card>
            </v-col>
          </template>
        </draggable>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useTitle } from '@vueuse/core'
import { useUql, QueryPart } from '@/use/uql'
import { UseMetrics } from '@/metrics/use-metrics'
import { useDashQueryManager, useDashEntryManager, UseDashboard } from '@/metrics/use-dashboards'

// Components
import draggable from 'vuedraggable'
import DashQueryBuilder from '@/metrics/query/DashQueryBuilder.vue'
import DashGaugeCard from '@/metrics/DashGaugeCard.vue'
import DashGridEntry from '@/metrics/DashGridEntry.vue'

export default defineComponent({
  name: 'DashGrid',
  components: {
    draggable,
    DashQueryBuilder,
    DashGaugeCard,
    DashGridEntry,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    metrics: {
      type: Object as PropType<UseMetrics>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<UseDashboard>,
      required: true,
    },
    baseQuery: {
      type: String,
      required: true,
    },
  },

  setup(props, ctx) {
    useTitle(computed(() => `${props.dashboard.data!.name} | Metrics`))

    const uql = useUql()
    const tableQueryMan = useDashQueryManager(props.dashboard)
    const dashEntryMan = useDashEntryManager()

    const metricNames = computed((): string[] => {
      if (!props.dashboard.status.isResolved()) {
        return []
      }

      const names: string[] = []
      for (let entry of props.dashboard.entries) {
        for (let m of entry.metrics) {
          names.push(m.name)
        }
      }
      return names
    })

    watch(
      () => props.baseQuery,
      (query) => {
        uql.query = query
      },
      { immediate: true },
    )

    watch(
      () => uql.query,
      (query) => {
        ctx.emit('update:base-query', query)
      },
    )

    function onBaseQuery(index: number, queryParts: QueryPart[]) {
      // Use information from the first grid item.
      if (index === 0) {
        uql.syncParts(queryParts)
      }
    }

    return { uql, tableQueryMan, dashEntryMan, metricNames, onBaseQuery }
  },
})
</script>

<style lang="scss" scoped></style>
