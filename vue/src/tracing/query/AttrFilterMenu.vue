<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" v-bind="attrs" v-on="on">
        <span>{{ label }}</span>
        <v-icon v-if="showIcon" right class="ml-0">mdi-menu-down</v-icon>
      </v-btn>
    </template>

    <SearchableList
      :loading="suggestions.loading"
      :items="suggestions.filteredItems"
      :num-item="suggestions.items.length"
      :search-input.sync="suggestions.searchInput"
      return-object
      @input="addFilter($event.value, '=')"
    >
      <template #item="{ item }">
        <v-list-item-content>
          <v-list-item-title>
            {{ truncateMiddle(item.value, 60) }}
          </v-list-item-title>
        </v-list-item-content>

        <v-list-item-action class="my-0" @click.stop="addFilter(item.value, '!=')">
          <v-btn icon>
            <v-icon small>mdi-not-equal</v-icon>
          </v-btn>
        </v-list-item-action>
      </template>
    </SearchableList>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { AxiosParams } from '@/use/axios'
import { UseUql } from '@/use/uql'
import { useDataSource } from '@/use/datasource'

// Components
import SearchableList from '@/components/SearchableList.vue'

// Misc
import { truncateMiddle } from '@/util/string'

export default defineComponent({
  name: 'AttrFilterMenu',
  components: { SearchableList },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<AxiosParams>,
      required: true,
    },
    attrKey: {
      type: String,
      required: true,
    },
    label: {
      type: String,
      required: true,
    },
    showIcon: {
      type: Boolean,
      default: false,
    },
  },

  setup(props, ctx) {
    const { route } = useRouter()
    const menu = shallowRef(false)

    const suggestions = useDataSource(() => {
      if (!menu.value) {
        return null
      }

      const { projectId } = route.value.params
      return {
        url: `/internal/v1/tracing/${projectId}/attributes/${props.attrKey}`,
        params: {
          ...props.axiosParams,
        },
      }
    })

    function addFilter(attrValue: string, op: string) {
      const editor = props.uql.createEditor()
      editor.where(props.attrKey, op, attrValue)
      props.uql.commitEdits(editor)

      menu.value = false
      ctx.emit('change')
    }

    return { menu, suggestions, addFilter, truncateMiddle }
  },
})
</script>

<style lang="scss" scoped></style>
