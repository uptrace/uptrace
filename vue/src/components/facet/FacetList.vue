<template>
  <div>
    <template v-if="!attrs.status.hasData()">
      <v-skeleton-loader v-for="i in 5" :key="i" class="mx-auto" type="card"></v-skeleton-loader>
    </template>

    <template v-else-if="!sortedCategories.length">
      <v-container fluid class="fill-height">
        <v-row>
          <v-col class="text-center">
            <v-card flat>
              <div class="mb-4">
                <v-icon size="48">mdi-magnify</v-icon>
              </div>

              <p class="text-body-1 text--secondary">
                There are no matching attributes.<br />
                Try to change filters.
              </p>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </template>

    <v-container v-else fluid class="py-6">
      <v-row>
        <v-col>
          <v-text-field
            v-model="searchInput"
            label="Filter attributes"
            outlined
            dense
            hide-details="auto"
            autofocus
            clearable
          ></v-text-field>
        </v-col>
      </v-row>

      <v-row v-for="category in sortedCategories" :key="category">
        <v-col>
          <FacetCategory
            :component="component"
            :axios-params="axiosParams"
            :filters-state="filtersState"
            :category="category"
            :items="categories[category]"
            :expanded="isExpanded(category)"
            @update:filter="updateQuery($event)"
            @update:pinned="attrs.reload()"
            @click:close="$emit('input', false)"
          />
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script lang="ts">
import { filter as fuzzyFilter } from 'fuzzaldrin-plus'
import { defineComponent, shallowRef, computed, onBeforeMount, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { UseUql } from '@/use/uql'
import { useDataSource, Item as BaseItem } from '@/use/datasource'
import { Category, Filter } from '@/components/facet/types'

// Components
import FacetCategory from '@/components/facet/FacetCategory.vue'

// Utilities
import { extractFilterState } from '@/components/facet/lexer'
import { AttrKey } from '@/models/otel'
import { quote, escapeRe } from '@/util/string'

type KVs = Record<string, string[]>

interface Item extends BaseItem {
  pinned: boolean
}

export default defineComponent({
  name: 'FacetList',
  components: { FacetCategory },

  props: {
    component: {
      type: String,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    axiosParams: {
      type: undefined as unknown as PropType<Record<string, any> | null>,
      required: true,
    },
    attrPrefix: {
      type: String,
      default: '',
    },
  },

  setup(props) {
    const route = useRoute()
    const searchInput = shallowRef('')

    const attrs = useDataSource<Item>(() => {
      if (!props.axiosParams) {
        return props.axiosParams
      }

      const { projectId } = route.value.params
      return {
        url: `/api/v1/${props.component}/${projectId}/attr-keys?kind=text`,
        params: props.axiosParams,
      }
    })

    const categories = computed((): Record<string, Item[]> => {
      if (searchInput.value) {
        const items = fuzzyFilter(attrs.items, searchInput.value, { key: 'text' })
        return { [Category.Found]: items }
      }

      if (attrs.items.length <= 10) {
        return { [Category.All]: attrs.items }
      }

      const categories: Record<string, Item[]> = {}
      const pinnedCategory = []

      for (let item of attrs.items) {
        if (item.pinned) {
          pinnedCategory.push(item)
          continue
        }

        const categoryName = attrCategory(item.value)
        let category = categories[categoryName]
        if (!category) {
          category = []
          categories[categoryName] = category
        }
        category.push(item)
      }

      if (pinnedCategory.length) {
        categories[Category.Pinned] = pinnedCategory
      }

      return categories
    })

    const sortedCategories = computed(() => {
      const sorted = Object.keys(categories.value).sort()

      const i = sorted.indexOf(Category.Pinned)
      if (i > 0) {
        sorted.splice(i, 1)
        sorted.unshift(Category.Pinned)
      }

      return sorted
    })

    const filtersState = computed((): KVs => {
      const kvs: KVs = {}
      for (let part of props.uql.parts) {
        const state = extractFilterState(part.query)
        if (state) {
          kvs[state.attr] = state.values
        }
      }
      return kvs
    })

    onBeforeMount(() => {
      const { envs, services } = route.value.query
      const items = [
        { attr: AttrKey.deploymentEnvironment, value: envs },
        { attr: AttrKey.service, value: services },
      ]

      for (let item of items) {
        if (!item.value) {
          continue
        }
        if (Array.isArray(item.value)) {
          updateQuery({
            attr: item.attr,
            op: 'in',
            value: item.value as string[],
          })
        } else if (item.value) {
          updateQuery({
            attr: item.attr,
            op: '=',
            value: [item.value],
          })
        }
      }
    })

    function attrCategory(attr: string): string {
      switch (attr) {
        case AttrKey.service:
          return AttrKey.service
      }

      const index = attr.indexOf('.')
      if (index >= 0) {
        return attr.slice(0, index)
      }
      return Category.Other
    }

    function hasActiveFilters(items: Item[]): boolean {
      return items.filter((item) => filtersState.value[item.value]).length > 0
    }

    function updateQuery(filter: Filter) {
      let { attr, op, value } = filter
      if (props.attrPrefix) {
        attr = props.attrPrefix + attr
      }

      const editor = props.uql.createEditor()
      const re = new RegExp(`^where\\s+${escapeRe(attr)}\\s+(=|in|like|not\\s+like)\\s+`, 'i')

      if (!value.length) {
        editor.remove(re)
        props.uql.query = editor.toString()
        return
      }

      let query: string
      if (value.length === 1) {
        query = `where ${attr} ${op} ${quote(value[0])}`
      } else {
        const values = value.map((value) => quote(value)).join(', ')
        query = `where ${attr} in (${values})`
      }

      editor.replaceOrPush(re, query)
      props.uql.query = editor.toString()
    }

    function isExpanded(category: string): boolean {
      if (sortedCategories.value.length === 1) {
        return true
      }
      if (searchInput.value && category === Category.Found) {
        return true
      }
      return category === Category.Pinned || hasActiveFilters(categories.value[category])
    }

    return {
      searchInput,
      attrs,

      categories,
      sortedCategories,
      filtersState,

      hasActiveFilters,
      updateQuery,
      isExpanded,
    }
  },
})
</script>

<style lang="scss" scoped></style>
