<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text :disabled="disabled" class="v-btn--filter" v-bind="attrs" v-on="on">
        {{ attrKey }}
      </v-btn>
    </template>

    <SearchableList
      :loading="suggestions.loading"
      :items="suggestions.filteredItems"
      :num-item="suggestions.items.length"
      :search-input.sync="suggestions.searchInput"
      return-object
      @input="whereEqual($event)"
    >
      <template #item="{ item }">
        <v-list-item-content>
          <v-list-item-title>{{ item.text }}</v-list-item-title>
        </v-list-item-content>

        <v-list-item-action class="my-0" @click.stop="whereNotEqual(item)">
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
import { useSuggestions } from '@/use/suggestions'
import { UseUql } from '@/use/uql'

// Components
import SearchableList from '@/components/SearchableList.vue'

// Utilities
import { quote } from '@/util/string'

interface WhereSuggestion {
  text: string
  key: string
  value: string
}

export default defineComponent({
  name: 'DashWhereMenu',
  components: { SearchableList },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    attrKey: {
      type: String,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
  },

  setup(props) {
    const { route } = useRouter()
    const menu = shallowRef(false)

    const suggestions = useSuggestions(() => {
      if (!menu.value) {
        return null
      }

      const { projectId } = route.value.params
      return {
        url: `/api/v1/metrics/${projectId}/where`,
        params: {
          ...props.axiosParams,
          query: props.uql.query,
          attr: props.attrKey,
        },
      }
    })

    function whereEqual(suggestion: WhereSuggestion) {
      where(suggestion, '=')
    }

    function whereNotEqual(suggestion: WhereSuggestion) {
      where(suggestion, '!=')
    }

    function where(suggestion: WhereSuggestion, op: string) {
      const editor = props.uql.createEditor()
      editor.where(suggestion.key, op, suggestion.value)
      props.uql.commitEdits(editor)

      menu.value = false
    }

    return { menu, suggestions, whereEqual, whereNotEqual, quote }
  },
})
</script>

<style lang="scss" scoped></style>
