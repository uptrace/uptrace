<template>
  <div>
    <v-simple-table :dense="dense" class="v-data-table--narrow">
      <colgroup>
        <col v-if="hasGroupName" class="target" />
        <col v-if="showSystemColumn" />
        <col v-for="col in customColumns" :key="col.name" />
        <col />
      </colgroup>

      <thead v-if="items.length" class="v-data-table-header">
        <tr>
          <th v-if="hasGroupName" class="target text-no-wrap">
            <span>Group Name</span>
          </th>
          <ThOrder v-if="showSystemColumn" :value="xkey.spanSystem" :order="order">System</ThOrder>
          <ThOrder
            v-for="col in customColumns"
            :key="col.name"
            :value="col.name"
            :order="order"
            :align="col.isNum ? 'center' : 'start'"
          >
            <span>{{ columnHeader(col) }}</span>
          </ThOrder>
          <ThOrder v-if="hasTimeColumn" :value="`max(${xkey.spanTime})`" :order="order">
            Time
          </ThOrder>
          <th></th>
        </tr>
      </thead>

      <thead v-show="loading">
        <tr class="v-data-table__progress">
          <th colspan="99" class="column">
            <v-progress-linear height="2" absolute indeterminate />
          </th>
        </tr>
      </thead>

      <tbody v-if="!items.length">
        <tr class="v-data-table__empty-wrapper">
          <td colspan="99" class="py-16">
            <div class="mb-4">There are no matching groups. Try to change filters.</div>
            <v-btn :to="{ name: 'TracingHelp' }">
              <v-icon left>mdi-help-circle-outline</v-icon>
              <span>Help</span>
            </v-btn>
          </td>
        </tr>
      </tbody>

      <tbody>
        <template v-for="item in items">
          <tr
            :key="item[xkey.itemId]"
            class="cursor-pointer"
            @click="groupViewer.toggle(item[xkey.itemId])"
          >
            <td v-if="hasGroupName" class="target">
              <span>{{ itemName(item) }}</span>
            </td>
            <td v-if="showSystemColumn">
              <router-link :to="systemRoute(item)" @click.native.stop>{{
                item[xkey.spanSystem]
              }}</router-link>
            </td>
            <td v-for="col in customColumns" :key="col.name">
              <div v-if="col.isNum" class="d-flex align-center justify-center">
                <LoadGroupSparkline
                  v-if="plotColumns.indexOf(col.name) >= 0"
                  :axios-params="axiosParams"
                  :where="groupBasedWhere(item)"
                  :column="col.name"
                  class="mr-2"
                />
                <XNum :value="item[col.name]" :name="col.name" />
              </div>

              <span v-else>{{ item[col.name] }}</span>
            </td>
            <td v-if="hasTimeColumn" class="text-no-wrap">
              <slot name="time" :item="item">
                <XDate :date="item[`max(${xkey.spanTime})`]" format="relative" />
              </slot>
            </td>
            <td class="text-center text-no-wrap">
              <v-btn
                icon
                title="Filter spans for this group"
                :to="exploreRoute(item)"
                exact
                @click.native.stop
              >
                <v-icon>mdi-filter-variant</v-icon>
              </v-btn>

              <v-btn
                v-if="groupViewer.visible(item[xkey.itemId])"
                icon
                title="Hide spans"
                @click.stop="groupViewer.hide(item[xkey.itemId])"
              >
                <v-icon size="30">mdi-chevron-up</v-icon>
              </v-btn>
              <v-btn
                v-else
                icon
                title="View spans"
                @click.stop="groupViewer.show(item[xkey.itemId])"
              >
                <v-icon size="30">mdi-chevron-down</v-icon>
              </v-btn>
            </td>
          </tr>
          <tr
            v-if="groupViewer.visible(item[xkey.itemId])"
            :key="`${item[xkey.itemId]}-spans`"
            class="v-data-table__expanded v-data-table__expanded__content"
          >
            <td colspan="99" class="px-6 pt-3 pb-4">
              <SpanListInline
                :date-range="dateRange"
                :systems="systems"
                :uql="uql"
                :is-event="isEventSystem(item[xkey.spanSystem])"
                :axios-params="axiosParams"
                :where="groupBasedWhere(item)"
                :span-list-route="spanListRoute"
                :group-list-route="groupListRoute"
              />
            </td>
          </tr>
        </template>
      </tbody>
    </v-simple-table>
  </div>
</template>

<script lang="ts">
import { omit, truncate } from 'lodash'
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseOrder } from '@/use/order'
import { UseSystems } from '@/use/systems'
import { ExploreItem, ColumnInfo } from '@/tracing/use-span-explore'
import { createUqlEditor, UseUql } from '@/use/uql'

// Components
import ThOrder from '@/components/ThOrder.vue'
import LoadGroupSparkline from '@/tracing/LoadGroupSparkline.vue'
import SpanListInline from '@/tracing/SpanListInline.vue'

// Utilities
import { xkey, isEventSystem, isDummySystem } from '@/models/otelattr'
import { quote } from '@/util/string'

// Styles
import 'vuetify/src/components/VDataTable/VDataTable.sass'

export default defineComponent({
  name: 'GroupsTable',
  components: {
    ThOrder,
    LoadGroupSparkline,
    SpanListInline,
  },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      default: undefined,
    },
    loading: {
      type: Boolean,
      required: true,
    },
    items: {
      type: Array as PropType<ExploreItem[]>,
      required: true,
    },
    columns: {
      type: Array as PropType<ColumnInfo[]>,
      required: true,
    },
    groupColumns: {
      type: Array as PropType<ColumnInfo[]>,
      required: true,
    },
    plotColumns: {
      type: Array as PropType<string[]>,
      required: true,
    },
    order: {
      type: Object as PropType<UseOrder>,
      required: true,
    },
    dense: {
      type: Boolean,
      default: false,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      default: undefined,
    },
    spanListRoute: {
      type: String,
      default: 'SpanList',
    },
    groupListRoute: {
      type: String,
      default: 'SpanGroupList',
    },
  },

  setup(props) {
    const { route } = useRouter()
    const groupViewer = useGroupViewer()

    const hasGroupName = computed((): boolean => {
      return hasColumn(xkey.spanName)
    })

    const showSystemColumn = computed((): boolean => {
      return isDummySystem(props.systems.activeSystem) && hasColumn(xkey.spanSystem)
    })

    const hasTimeColumn = computed(() => {
      return hasColumn(`max(${xkey.spanTime})`)
    })

    const customColumns = computed(() => {
      const blacklist = [xkey.spanSystem, `max(${xkey.spanTime})`]
      if (hasGroupName.value) {
        blacklist.push(xkey.spanGroupId, xkey.spanName)
      }

      const columns = props.columns.filter((col) => {
        return blacklist.indexOf(col.name) === -1
      })

      return columns
    })

    function hasColumn(name: string): boolean {
      if (!props.items.length) {
        return false
      }
      const item = props.items[0]
      return name in item
    }

    function exploreRoute(item: ExploreItem) {
      const editor = props.uql
        ? props.uql.createEditor()
        : createUqlEditor().exploreAttr(xkey.spanGroupId)

      for (let col of props.groupColumns) {
        const value = item[col.name]
        editor.where(col.name, '=', value)
      }

      return {
        name: props.spanListRoute,
        query: {
          ...omit(route.value.query, 'column'),
          ...props.systems.axiosParams(),
          query: editor.toString(),
        },
      }
    }

    function columnHeader(col: ColumnInfo) {
      switch (col.name) {
        case xkey.spanErrorCount:
          return 'errors'
        case xkey.spanErrorPct:
          return 'err%'
      }

      const m = col.name.match(/^([0-9a-z]+)\(span\.duration\)$/)
      if (m) {
        return m[1]
      }

      const spanPrefix = 'span.'
      if (col.name.startsWith(spanPrefix)) {
        return col.name.slice(spanPrefix.length)
      }

      return col.name
    }

    function groupBasedWhere(item: ExploreItem) {
      const ss = []
      for (let col of props.groupColumns) {
        const value = item[col.name]
        ss.push(`${col.name} = ${quote(value)}`)
      }
      return `where ${ss.join(' AND ')}`
    }

    function systemRoute(item: any) {
      return {
        query: {
          ...route.value.query,
          system: item[xkey.spanSystem],
        },
      }
    }

    return {
      xkey,
      groupViewer,

      hasGroupName,
      showSystemColumn,
      hasTimeColumn,
      customColumns,

      isEventSystem,
      columnHeader,
      exploreRoute,
      groupBasedWhere,
      systemRoute,
      itemName,
    }
  },
})

function useGroupViewer() {
  const activeItemId = shallowRef<number>()

  function visible(itemId: number): boolean {
    return activeItemId.value === itemId
  }

  function toggle(itemId: number) {
    if (visible(itemId)) {
      hide(itemId)
    } else {
      show(itemId)
    }
  }

  function show(itemId: number) {
    activeItemId.value = itemId
  }

  function hide(_itemId: number) {
    activeItemId.value = undefined
  }

  return { visible, show, hide, toggle }
}

function itemName(item: Record<string, any>, maxLength = 120): string {
  const eventName = item[xkey.spanEventName]
  if (eventName) {
    return truncate(eventName, { length: maxLength })
  }

  const name = item[xkey.spanName]
  return truncate(name, { length: maxLength })
}
</script>

<style lang="scss" scoped></style>
