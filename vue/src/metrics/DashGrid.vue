<template>
  <v-card flat color="transparent" :max-width="dashboard.gridMaxWidth" class="mx-auto">
    <v-container fluid>
      <v-row v-if="gridRows.length" dense>
        <v-col>
          <v-card outlined rounded="lg" class="py-2 px-4">
            <DashQueryBuilder
              :date-range="dateRange"
              :metrics="gridMetrics"
              :uql="uql"
              class="mb-1"
            >
              <template #prepend-actions>
                <v-btn
                  v-if="isGridQueryDirty"
                  :loading="dashMan.pending"
                  small
                  depressed
                  class="mr-4"
                  @click="saveGridQuery()"
                >
                  <v-icon small left>mdi-check</v-icon>
                  <span>Save</span>
                </v-btn>
              </template>
            </DashQueryBuilder>
          </v-card>
        </v-col>
      </v-row>

      <v-row v-if="!gridRows.length" dense>
        <v-col v-for="i in 6" :key="i" cols="6">
          <v-skeleton-loader type="image" boilerplate></v-skeleton-loader>
        </v-col>
      </v-row>

      <v-row v-for="row in gridRows" :key="row.id" dense>
        <v-col>
          <GridRowCard :row="row" @change="$emit('change')">
            <template #item="{ attrs, on }">
              <GridItemAny
                :date-range="dateRange"
                :dashboard="dashboard"
                v-bind="attrs"
                :grid-query="gridQuery"
                v-on="on"
              />
            </template>
          </GridRowCard>
        </v-col>
      </v-row>
    </v-container>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, computed, watch, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useUql, useQueryStore, provideQueryStore } from '@/use/uql'
import { useDashboardManager } from '@/metrics/use-dashboards'

// Components
import DashQueryBuilder from '@/metrics/query/DashQueryBuilder.vue'
import GridRowCard from '@/metrics/GridRowCard.vue'
import GridItemAny from '@/metrics/GridItemAny.vue'

// Misc
import { Dashboard, GridRow, GridItemType, DashKind } from '@/metrics/types'

export default defineComponent({
  name: 'DashdGrid',
  components: {
    DashQueryBuilder,
    GridRowCard,
    GridItemAny,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    dashboard: {
      type: Object as PropType<Dashboard>,
      required: true,
    },
    gridRows: {
      type: Array as PropType<GridRow[]>,
      required: true,
    },
    gridMetrics: {
      type: Array as PropType<string[]>,
      required: true,
    },
    gridQuery: {
      type: String,
      default: '',
    },
  },

  setup(props, ctx) {
    const uql = useUql()
    provideQueryStore(useQueryStore(uql))
    watch(
      () => props.gridQuery,
      (query) => {
        uql.query = query
      },
      { immediate: true },
    )

    const dashMan = useDashboardManager()
    const isGridQueryDirty = computed(() => {
      return uql.query !== props.gridQuery
    })
    function saveGridQuery() {
      dashMan.updateGrid({ id: props.dashboard.id, gridQuery: uql.query }).then(() => {
        ctx.emit('change')
      })
    }

    return {
      DashKind,
      GridItemType,

      uql,

      dashMan,
      isGridQueryDirty,
      saveGridQuery,
    }
  },
})
</script>

<style lang="scss" scoped></style>
