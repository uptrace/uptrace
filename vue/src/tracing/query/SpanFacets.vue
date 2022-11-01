<template>
  <XPlaceholder>
    <template v-if="!attrs.status.hasData()" #placeholder>
      <v-skeleton-loader v-for="i in 5" :key="i" class="mx-auto" type="card"></v-skeleton-loader>
    </template>

    <template v-else-if="!sortedCategories.length" #placeholder>
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

    <v-container fluid>
      <v-card flat>
        <SpanFacetsCategory
          v-for="category in sortedCategories"
          :key="category"
          :axios-params="axiosParams"
          :filters-state="filtersState"
          :category="category"
          :items="categories[category]"
          :expanded="category === CATEGORY_CORE || hasActiveFilters(categories[category])"
          @update:filter="updateQuery($event.attr, $event.value)"
          @click:add-query="addQuery($event)"
          @click:close="$emit('input', false)"
        />
      </v-card>
    </v-container>
  </XPlaceholder>
</template>

<script lang="ts">
import { defineComponent, computed, onBeforeMount, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { UseUql } from '@/use/uql'
import { useSpanAttrs, Item, CATEGORY_CORE } from '@/tracing/query/use-span-facets'

// Components
import SpanFacetsCategory from '@/tracing/query/SpanFacetsCategory.vue'

// Utilities
import { extractFilterState } from '@/tracing/query/lexer'
import { AttrKey } from '@/models/otel'
import { quote, escapeRe } from '@/util/string'

type KVs = Record<string, string[]>

const CORE_ATTRS = [
  AttrKey.spanStatusCode,
  AttrKey.spanKind,
  AttrKey.deploymentEnvironment,
  AttrKey.service,
  AttrKey.hostName,
  AttrKey.rpcMethod,
  AttrKey.httpMethod,
  AttrKey.httpStatusCode,
  AttrKey.dbOperation,
  AttrKey.dbSqlTables,
  AttrKey.logSeverity,
  AttrKey.logSource,
  AttrKey.logFilePath,
  AttrKey.logFileName,
  AttrKey.exceptionType,
  AttrKey.codeFilepath,
  AttrKey.codeFunction,
]

export default defineComponent({
  name: 'SpanFacets',
  components: { SpanFacetsCategory },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    axiosParams: {
      type: undefined as unknown as PropType<Record<string, any> | null>,
      required: true,
    },
  },

  setup(props, ctx) {
    const route = useRoute()

    const attrs = useSpanAttrs(() => props.axiosParams)

    const categories = computed(() => {
      const categories: Record<string, Item[]> = {}
      const coreCategory = []

      for (let item of attrs.items) {
        const index = CORE_ATTRS.indexOf(item.text as AttrKey)
        if (index >= 0) {
          coreCategory.push(item)
          continue
        }

        const categoryName = attrCategory(item.text)
        let category = categories[categoryName]
        if (!category) {
          category = []
          categories[categoryName] = category
        }
        category.push(item)
      }

      if (coreCategory.length) {
        coreCategory.sort((a, b) => {
          return CORE_ATTRS.indexOf(a.text as AttrKey) - CORE_ATTRS.indexOf(b.text as AttrKey)
        })
        categories[CATEGORY_CORE] = coreCategory
      }

      return categories
    })

    const sortedCategories = computed(() => {
      const sorted = Object.keys(categories.value).sort()

      const i = sorted.indexOf(CATEGORY_CORE)
      if (i > 0) {
        sorted.splice(i, 1)
        sorted.unshift(CATEGORY_CORE)
      }

      return sorted
    })

    const filtersState = computed((): KVs => {
      const kvs: KVs = {}
      for (let part of props.uql.parts) {
        const state = extractFilterState(part.query)
        if (state !== null) {
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
          updateQuery(item.attr, item.value as string[])
        } else if (item.value === 'string') {
          updateQuery(item.attr, [item.value])
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
      return 'Other'
    }

    function hasActiveFilters(items: Item[]): boolean {
      return items.filter((item) => filtersState.value[item.text]).length > 0
    }

    function addQuery(query: string) {
      const editor = props.uql.createEditor()
      editor.add(query)
      props.uql.query = editor.toString()

      // Close drawer.
      ctx.emit('input', false)
    }

    function updateQuery(attr: string, values: string[]) {
      const editor = props.uql.createEditor()
      const re = new RegExp(`^where\\s+${escapeRe(attr)}\\s+(=|in)\\s+`, 'i')

      if (!values.length) {
        editor.remove(re)
        props.uql.query = editor.toString()
        return
      }

      let query: string
      if (values.length === 1) {
        const value = values[0]
        query = `where ${attr} = ${quote(value)}`
      } else {
        const value = values.map((value) => quote(value)).join(', ')
        query = `where ${attr} in (${value})`
      }

      editor.replaceOrPush(re, query)
      props.uql.query = editor.toString()
    }

    return {
      CATEGORY_CORE,
      attrs,

      categories,
      sortedCategories,
      filtersState,

      hasActiveFilters,
      addQuery,
      updateQuery,
    }
  },
})
</script>

<style lang="scss" scoped></style>
