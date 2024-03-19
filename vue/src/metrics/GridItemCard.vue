<template>
  <v-card flat outlined>
    <v-toolbar :color="toolbarColor" flat height="36">
      <v-toolbar-title class="d-flex align-center flex-grow-1 drag-handle cursor-move">
        <v-tooltip max-width="500" bottom :disabled="!Boolean(columnError || gridItem.description)">
          <template #activator="{ on, attrs }">
            <span
              class="text-subtitle-2 font-weight-regular"
              :class="{ 'red--text text--darken-1': Boolean(columnError) }"
              v-bind="attrs"
              v-on="on"
            >
              {{ gridItem.title }}
            </span>
          </template>
          <div>{{ columnError || gridItem.description }}</div>
        </v-tooltip>
      </v-toolbar-title>

      <v-menu v-model="menu" offset-y>
        <template #activator="{ on: onMenu, attrs }">
          <v-btn
            :loading="gridItemMan.pending"
            icon
            :disabled="readonly"
            v-bind="attrs"
            v-on="onMenu"
          >
            <v-icon>mdi-dots-vertical</v-icon>
          </v-btn>
        </template>
        <v-list>
          <v-list-item v-if="expandable" @click="dialog = true">
            <v-list-item-icon>
              <v-icon>mdi-eye</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title>View</v-list-item-title>
            </v-list-item-content>
          </v-list-item>

          <v-list-item @click="$emit('click:edit', gridItem)">
            <v-list-item-icon>
              <v-icon>mdi-pencil</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title>Edit</v-list-item-title>
            </v-list-item-content>
          </v-list-item>

          <v-list-item v-if="routeForNewMonitor" :to="routeForNewMonitor">
            <v-list-item-icon>
              <v-icon>mdi-radar</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title>Monitor</v-list-item-title>
            </v-list-item-content>
          </v-list-item>

          <v-list-item @click="del">
            <v-list-item-icon>
              <v-icon>mdi-delete</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title>Delete</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-toolbar>

    <slot
      :height="height - 44"
      :wide="false"
      :on="{
        error($event) {
          columnError = $event
        },
      }"
    />

    <DialogCard v-model="dialog" max-width="1200" :title="gridItem.title">
      <template v-if="gridItem.description" #toolbar-append>
        <v-tooltip bottom>
          <template #activator="{ on, attrs }">
            <v-toolbar-items v-bind="attrs" v-on="on">
              <v-icon class="ml-2">mdi-information-outline</v-icon>
            </v-toolbar-items>
          </template>
          <span>{{ gridItem.description }}</span>
        </v-tooltip>
      </template>
      <div class="pa-2 pb-4">
        <slot :height="500" wide :on="{}" />
      </div>
    </DialogCard>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { useGridItemManager } from '@/metrics/use-dashboards'

// Components
import DialogCard from '@/components/DialogCard.vue'

// Misc
import { GridItem, GridItemType } from '@/metrics/types'

export default defineComponent({
  name: 'GridItemCard',
  components: { DialogCard },

  props: {
    gridItem: {
      type: Object as PropType<GridItem>,
      required: true,
    },
    height: {
      type: Number,
      default: 0,
    },
    readonly: {
      type: Boolean,
      default: false,
    },
    expandable: {
      type: Boolean,
      default: false,
    },
    toolbarColor: {
      type: String,
      default: 'bg--primary',
    },
  },

  setup(props, ctx) {
    const menu = shallowRef(false)
    const dialog = shallowRef(false)

    const columnError = shallowRef(false)
    const routeForNewMonitor = computed(() => {
      switch (props.gridItem.type) {
        case GridItemType.Chart:
        case GridItemType.Table:
          return {
            name: 'MonitorMetricNew',
            query: {
              name: props.gridItem.title,
              metric: props.gridItem.params.metrics.map((m) => m.name),
              alias: props.gridItem.params.metrics.map((m) => m.alias),
              query: props.gridItem.params.query,
              columns: JSON.stringify(props.gridItem.params.columnMap),
            },
          }
        default:
          return undefined
      }
    })

    const gridItemMan = useGridItemManager()

    function del() {
      gridItemMan.delete(props.gridItem).then(() => {
        ctx.emit('change')
      })
    }

    return {
      GridItemType,

      menu,
      dialog,

      columnError,
      routeForNewMonitor,

      gridItemMan,
      del,
    }
  },
})
</script>

<style lang="scss" scoped>
.v-toolbar ::v-deep .v-toolbar__content {
  padding: 4px 8px;
}
</style>
