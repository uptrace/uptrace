<template>
  <div class="d-inline-block">
    <v-btn small plain @click="drawer = true">
      <v-icon left>mdi-dock-left</v-icon>
      <span>Views</span>
    </v-btn>

    <v-navigation-drawer
      v-model="drawer"
      v-click-outside="{
        handler: onClickOutside,
        closeConditional,
      }"
      app
      temporary
      stateless
      width="500"
    >
      <v-container fluid class="py-6">
        <v-row no-gutters>
          <v-col>
            <v-text-field
              v-model="searchInput"
              label="Filter views"
              outlined
              dense
              hide-details="auto"
              autofocus
              clearable
            ></v-text-field>
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <v-divider />
          </v-col>
        </v-row>
        <v-row>
          <v-col>
            <SavedViewsList
              :loading="views.loading"
              :items="filteredViews"
              :views="views"
              @click:item="drawer = false"
            />
          </v-col>
        </v-row>
      </v-container>
    </v-navigation-drawer>
  </div>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { SavedView, UseSavedViews } from '@/tracing/views/use-saved-views'

// Components
import SavedViewsList from '@/tracing/views/SavedViewsList.vue'

export default defineComponent({
  name: 'ListViewDrawer',
  components: { SavedViewsList },

  props: {
    views: {
      type: Object as PropType<UseSavedViews>,
      required: true,
    },
  },

  setup(props) {
    const searchInput = shallowRef('')
    const drawer = shallowRef(false)

    const filteredViews = computed((): SavedView[] => {
      if (!searchInput.value) {
        return props.views.items
      }
      return fuzzyFilter(props.views.items, searchInput.value, { key: 'name' })
    })

    function onClickOutside() {
      drawer.value = false
    }

    function closeConditional() {
      return drawer.value
    }

    return {
      searchInput,
      filteredViews,

      drawer,
      onClickOutside,
      closeConditional,
    }
  },
})
</script>

<style lang="scss" scoped></style>
