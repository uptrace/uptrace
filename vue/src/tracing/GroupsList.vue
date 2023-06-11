<template>
  <div>
    <v-row v-if="groups.length" dense align="center">
      <v-col cols="7" lg="8" xl="9">
        <slot name="actions">
          <v-slide-group
            v-if="systemFilterItems.length > 1"
            v-model="systemFilter"
            multiple
            center-active
            show-arrows
          >
            <v-slide-item
              v-for="(item, i) in systemFilterItems"
              :key="item.system"
              v-slot="{ active, toggle }"
              :value="item.system"
            >
              <v-btn
                :input-value="active"
                active-class="light-blue white--text"
                small
                depressed
                rounded
                :class="{ 'ml-1': i > 0 }"
                @click="toggle"
              >
                {{ item.system }} ({{ item.numGroup }})
              </v-btn>
            </v-slide-item>
          </v-slide-group>
        </slot>
      </v-col>

      <v-spacer />

      <v-col v-if="showPlottedColumnItems" cols="5" lg="4" xl="3">
        <v-select
          v-model="internalPlottedColumns"
          :items="plottableColumnItems"
          multiple
          dense
          solo
          flat
          background-color="grey lighten-4"
          hide-details="auto"
          @input="$emit('update:plotted-columns', $event)"
        ></v-select>
      </v-col>
    </v-row>

    <v-row dense>
      <v-col>
        <GroupsTable
          :date-range="dateRange"
          :systems="systems"
          :query="query"
          :loading="loading"
          :groups="pagedGroups"
          :columns="columns"
          :plottable-columns="plottableColumns"
          :plotted-columns="internalPlottedColumns"
          :order="order"
          :axios-params="axiosParams"
          :events-mode="eventsMode"
          :show-system="showSystem"
          :hide-actions="hideActions"
          @click:metrics="
            activeGroup = $event
            dialog = true
          "
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col>
        <XPagination :pager="pager" />
      </v-col>
    </v-row>

    <v-dialog v-model="dialog" max-width="1400">
      <v-card v-if="activeGroup" flat>
        <v-toolbar flat color="light-blue lighten-5">
          <v-toolbar-title>{{ activeGroup._query }}</v-toolbar-title>

          <v-spacer />

          <DateRangePicker :date-range="dateRange" :range-days="90" />

          <v-toolbar-items>
            <v-btn icon @click="dialog = false"><v-icon>mdi-close</v-icon></v-btn>
          </v-toolbar-items>
        </v-toolbar>

        <v-container fluid>
          <GroupMetrics
            :date-range="dateRange"
            :metrics="activeGroup.metrics"
            :where="activeGroup._query"
          />
        </v-container>
      </v-card>
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { orderBy } from 'lodash-es'
import { defineComponent, shallowRef, computed, watch, PropType } from 'vue'

// Composables
import { usePager } from '@/use/pager'
import { UseOrder } from '@/use/order'
import { UseDateRange } from '@/use/date-range'
import { Group, ColumnInfo } from '@/tracing/use-explore-spans'

// Components
import DateRangePicker from '@/components/date/DateRangePicker.vue'
import GroupsTable from '@/tracing/GroupsTable.vue'
import GroupMetrics from '@/metrics/GroupMetrics.vue'

// Utilities
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'GroupsList',
  components: { DateRangePicker, GroupsTable, GroupMetrics },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Array as PropType<string[]>,
      required: true,
    },
    query: {
      type: String,
      default: '',
    },
    loading: {
      type: Boolean,
      required: true,
    },
    groups: {
      type: Array as PropType<Group[]>,
      required: true,
    },
    columns: {
      type: Array as PropType<ColumnInfo[]>,
      required: true,
    },
    plottableColumns: {
      type: Array as PropType<ColumnInfo[]>,
      required: true,
    },
    plottedColumns: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    showPlottedColumnItems: {
      type: Boolean,
      default: false,
    },
    order: {
      type: Object as PropType<UseOrder>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      default: undefined,
    },
    eventsMode: {
      type: Boolean,
      default: false,
    },
    showSystem: {
      type: Boolean,
      default: false,
    },
    hideActions: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const dialog = shallowRef(false)
    const activeGroup = shallowRef<Group>()

    const systemFilter = shallowRef<string[]>([])
    const systemFilterItems = computed((): SystemFilter[] => {
      const filters = buildSystemFilters(props.groups)
      if (filters.length <= 5) {
        return filters
      }
      return buildTypeFilters(props.groups)
    })
    watch(systemFilterItems, () => {
      systemFilter.value = []
    })

    const filteredGroups = computed(() => {
      if (!systemFilter.value.length) {
        return props.groups
      }

      return props.groups.filter((group) => {
        const system = group[AttrKey.spanSystem]
        if (!system) {
          return true
        }

        for (let needle of systemFilter.value) {
          if (system.startsWith(needle)) {
            return true
          }
        }
        return false
      })
    })

    const pager = usePager({ perPage: 15 })
    const pagedGroups = computed((): Group[] => {
      const pagedGroups = filteredGroups.value.slice(pager.pos.start, pager.pos.end)
      return pagedGroups
    })
    watch(
      () => filteredGroups.value.length,
      (numItem) => {
        pager.numItem = numItem
        ctx.emit('update:num-group', numItem)
      },
      { immediate: true },
    )

    const internalPlottedColumns = shallowRef<string[]>()
    watch(
      () => props.plottedColumns,
      (plottedColumns) => {
        internalPlottedColumns.value = plottedColumns
      },
      { immediate: true },
    )
    const plottableColumnItems = computed(() => {
      const items = props.plottableColumns.map((col) => {
        return { text: col.name, value: col.name }
      })
      return items
    })

    return {
      dialog,
      activeGroup,

      pagedGroups,
      pager,

      internalPlottedColumns,
      plottableColumnItems,

      systemFilter,
      systemFilterItems,
    }
  },
})

interface SystemFilter {
  system: string
  numGroup: number
}

function buildSystemFilters(groups: Group[]) {
  const systemMap: Record<string, SystemFilter> = {}

  for (let group of groups) {
    const system = group[AttrKey.spanSystem]
    if (!system) {
      continue
    }

    let item = systemMap[system]
    if (!item) {
      item = {
        system,
        numGroup: 0,
      }
      systemMap[system] = item
    }
    item.numGroup++
  }

  const filters: SystemFilter[] = []

  for (let system in systemMap) {
    filters.push(systemMap[system])
  }

  orderBy(filters, 'system')
  return filters
}

function buildTypeFilters(groups: Group[]) {
  const systemMap: Record<string, SystemFilter> = {}

  for (let group of groups) {
    let system = group[AttrKey.spanSystem]
    if (!system) {
      continue
    }

    const i = system.indexOf(':')
    if (i >= 0) {
      system = system.slice(0, i)
    }

    let item = systemMap[system]
    if (!item) {
      item = {
        system,
        numGroup: 0,
      }
      systemMap[system] = item
    }
    item.numGroup++
  }

  const filters: SystemFilter[] = []

  for (let system in systemMap) {
    filters.push(systemMap[system])
  }

  orderBy(filters, 'system')
  return filters
}
</script>

<style lang="scss" scoped></style>
