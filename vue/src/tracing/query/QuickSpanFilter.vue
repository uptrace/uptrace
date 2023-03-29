<template>
  <v-autocomplete
    ref="autocomplete"
    v-model="activeValue"
    v-autowidth="{ minWidth: 40 }"
    :items="attrValues.filteredItems"
    :search-input.sync="attrValues.searchInput"
    no-filter
    placeholder="none"
    :prefix="`${name}: `"
    multiple
    clearable
    auto-select-first
    hide-details
    dense
    outlined
    background-color="light-blue lighten-5"
    class="v-select--fit"
  >
    <template #item="{ item, attrs }">
      <v-list-item
        v-bind="attrs"
        @click="
          activeValue = [item.value]
          autocomplete.blur()
        "
      >
        <v-list-item-action class="my-0 mr-4">
          <v-checkbox
            :input-value="activeValue.indexOf(item.value) >= 0"
            dense
            @click.stop="toggleValue(item.value)"
          ></v-checkbox>
        </v-list-item-action>
        <v-list-item-content>
          <v-list-item-title>{{ item.value }}</v-list-item-title>
        </v-list-item-content>
        <v-list-item-action class="my-0">
          <v-list-item-action-text><XNum :value="item.count" /></v-list-item-action-text>
        </v-list-item-action>
      </v-list-item>
    </template>
    <template #selection="{ index, item }">
      <div v-if="index === 2" class="v-select__selection">, {{ value.length - 2 }} more</div>
      <div v-else-if="index < 2" class="v-select__selection">
        {{ withComma(item, index) }}
      </div>
    </template>
    <template #no-data>
      <div>
        <v-list-item>
          <v-list-item-content>
            <v-list-item-title class="text-subtitle-1 font-weight-regular">
              To start filtering, set <code>{{ attrKey }}</code> attribute.
            </v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <div class="my-4 d-flex justify-center">
          <v-btn
            href="https://uptrace.dev/get/get-started.html#resource-attributes"
            target="_blank"
            color="primary"
          >
            <span>Open documentation</span>
            <v-icon right>mdi-open-in-new</v-icon>
          </v-btn>
        </div>
      </div>
    </template>
  </v-autocomplete>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { UseDateRange } from '@/use/date-range'
import { UseUql } from '@/use/uql'
import { useDataSource, Item } from '@/use/datasource'

// Utilities
import { truncateMiddle } from '@/util/string'
import { extractFilterState } from '@/components/facet/lexer'
import { quote, escapeRe } from '@/util/string'

export default defineComponent({
  name: 'QuickSpanFilter',

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    name: {
      type: String,
      required: true,
    },
    attrKey: {
      type: String,
      required: true,
    },
  },

  setup(props, ctx) {
    const route = useRoute()
    const autocomplete = shallowRef()

    const attrValues = useDataSource(() => {
      const { projectId } = route.value.params
      return {
        url: `/api/v1/tracing/${projectId}/attr-values?attr_key=${props.attrKey}`,
        params: {
          ...props.dateRange.axiosParams(),
          ...props.uql.axiosParams(),
        },
        cache: true,
      }
    })

    const activeValue = computed({
      get() {
        for (let part of props.uql.parts) {
          const state = extractFilterState(part.query)
          if (!state) {
            continue
          }
          if (state.attr === props.attrKey) {
            return state.values
          }
        }
        return []
      },
      set(values: string[]) {
        const editor = props.uql.createEditor()
        const re = new RegExp(
          `^where\\s+${escapeRe(props.attrKey)}\\s+(=|in|like|not\\s+like)\\s+`,
          'i',
        )

        if (!values.length) {
          editor.remove(re)
          props.uql.query = editor.toString()
          return
        }

        let query: string
        if (values.length === 1) {
          query = `where ${props.attrKey} = ${quote(values[0])}`
        } else {
          query = `where ${props.attrKey} in (${values.join(', ')})`
        }

        editor.replaceOrPush(re, query)
        props.uql.query = editor.toString()
      },
    })

    function withComma(item: Item, index: number): string {
      const value = truncateMiddle(item.value, 20)
      if (index > 0) {
        return ', ' + value
      }
      return value
    }

    function toggleValue(value: string) {
      const values = activeValue.value.slice()

      const index = values.indexOf(value)
      if (index >= 0) {
        values.splice(index, 1)
      } else {
        values.push(value)
      }

      activeValue.value = values
    }

    return {
      autocomplete,
      activeValue,
      attrValues,
      withComma,
      toggleValue,
    }
  },
})
</script>

<style lang="scss" scoped></style>
