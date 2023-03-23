<template>
  <v-card :loading="loading" flat>
    <v-card-text v-if="!items.length" class="text-center">
      The list is empty.<br />
      Have you saved any views yet?
    </v-card-text>

    <v-list v-else class="py-0">
      <template v-for="(item, i) in items">
        <v-divider v-if="i > 0" :key="`divider-${item.id}`"></v-divider>

        <v-list-item
          :key="item.id"
          :to="{ name: item.route, params: item.params, query: item.query }"
          exact
          three-line
          @click="$emit('click:item', item)"
        >
          <v-list-item-content>
            <v-list-item-title>
              <template v-if="tabName(item)">
                <span>{{ tabName(item) }}</span>
                <span class="px-2">&gt;</span>
              </template>

              <template v-if="item.query.system">
                <span>{{ querySystem(item.query) }}</span>
                <span class="px-2">&gt;</span>
              </template>

              <span>{{ item.name }}</span>

              <v-btn
                v-if="item.pinned"
                :loading="pendingView?.id === item.id"
                icon
                title="Unpin view"
                class="ml-2"
                @click.stop.prevent="unpinView(item)"
              >
                <v-icon size="20" color="green darken-2">mdi-pin</v-icon>
              </v-btn>
              <v-btn
                v-else
                :loading="pendingView?.id === item.id"
                icon
                title="Pin view to the top"
                class="ml-2"
                @click.stop.prevent="pinView(item)"
              >
                <v-icon size="20">mdi-pin-outline</v-icon>
              </v-btn>

              <v-btn
                :loading="pendingView?.id === item.id"
                icon
                title="Delete view"
                @click.stop.prevent="deleteView(item)"
              >
                <v-icon size="20">mdi-delete-outline</v-icon>
              </v-btn>
            </v-list-item-title>
            <v-list-item-subtitle>{{ subtitle(item) }}</v-list-item-subtitle>
          </v-list-item-content>
        </v-list-item>
      </template>
    </v-list>
  </v-card>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { useSavedViewManager, SavedView, UseSavedViews } from '@/tracing/views/use-saved-views'

export default defineComponent({
  name: 'SavedViewList',

  props: {
    loading: {
      type: Boolean,
      required: true,
    },
    items: {
      type: Array as PropType<SavedView[]>,
      required: true,
    },
    editMode: {
      type: Boolean,
      default: false,
    },
    views: {
      type: Object as PropType<UseSavedViews>,
      required: true,
    },
  },

  setup(props) {
    const viewMan = useSavedViewManager()
    const pendingView = shallowRef<SavedView>()

    function pinView(view: SavedView) {
      pendingView.value = view
      viewMan.pin(view.id).finally(() => {
        pendingView.value = undefined
        props.views.reload()
      })
    }

    function unpinView(view: SavedView) {
      pendingView.value = view
      viewMan.unpin(view.id).finally(() => {
        pendingView.value = undefined
        props.views.reload()
      })
    }

    function deleteView(view: SavedView) {
      pendingView.value = view
      viewMan.del(view.id).finally(() => {
        pendingView.value = undefined
        props.views.reload()
      })
    }

    function tabName(item: SavedView): string {
      switch (item.route) {
        case 'SpanGroupList':
        case 'EventGroupList':
          return 'Groups'
        case 'SpanList':
          return 'Spans'
        case 'EventList':
          return 'Events'
        case 'SpanTimeseries':
        case 'EventTimeseries':
          return 'Timeseries'
        default:
          return ''
      }
    }

    function subtitle(item: SavedView): string {
      return item.query.query ?? ''
    }

    function querySystem(query: Record<string, any>): string {
      const system = query.system
      if (Array.isArray(system) && system.length) {
        return system[0]
      }
      return system
    }

    return {
      viewMan,
      pendingView,
      pinView,
      unpinView,
      deleteView,

      tabName,
      subtitle,
      querySystem,
    }
  },
})
</script>

<style lang="scss" scoped></style>
