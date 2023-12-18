<template>
  <div>
    <v-menu v-if="menuItems.length" v-model="menu" offset-y>
      <template #activator="{ on }">
        <v-btn icon v-on="on"><v-icon>mdi-dots-vertical</v-icon></v-btn>
      </template>

      <v-list dense>
        <v-list-item v-for="item in menuItems" :key="item.text" v-bind="item.attrs">
          <v-list-item-content class="text-body-2">
            <v-list-item-title>{{ item.text }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-menu>

    <AnyValue :value="attrValue" :name="attrKey" />
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'
import { createQueryEditor, injectQueryStore } from '@/use/uql'

// Utilities
import { isSpanSystem, AttrKey } from '@/models/otel'
import { truncateMiddle, quote } from '@/util/string'

export default defineComponent({
  name: 'SpanAttrValue',

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    system: {
      type: String,
      required: true,
    },
    groupId: {
      type: String,
      required: true,
    },
    attrKey: {
      type: String,
      required: true,
    },
    attrValue: {
      type: undefined,
      required: true,
    },
  },

  setup(props) {
    const menu = shallowRef(false)
    const { query, where } = injectQueryStore()

    const menuItems = computed(() => {
      const quotedValue = quote(truncateMiddle(props.attrValue, 60))

      const items = [
        {
          text: `Group by ${props.attrKey}`,
          attrs: createLink('SpanGroupList', {
            query: createQueryEditor()
              .exploreAttr(props.attrKey, isSpanSystem(props.system))
              .add(where.value)
              .where(AttrKey.spanGroupId, '=', props.groupId)
              .where(props.attrKey, 'exists')
              .toString(),
          }),
        },
      ]

      if (query.value) {
        items.push({
          text: `${props.attrKey} = ${quotedValue}`,
          attrs: createLink(undefined, {
            query: createQueryEditor()
              .add(query.value)
              .where(props.attrKey, '=', props.attrValue)
              .toString(),
          }),
        })
      } else {
        items.push({
          text: `Groups with ${props.attrKey} = ${quotedValue}`,
          attrs: createLink('SpanGroupList', {
            query: createQueryEditor()
              .exploreAttr(AttrKey.spanGroupId, isSpanSystem(props.system))
              .where(props.attrKey, '=', props.attrValue)
              .toString(),
          }),
        })
      }

      return items
    })

    function createLink(routeName: string | undefined, queryParams: Record<string, any> = {}) {
      return {
        to: {
          name: routeName,
          query: {
            ...props.dateRange.queryParams(),
            ...queryParams,
          },
        },
        exact: true,
      }
    }

    return { menu, menuItems }
  },
})
</script>

<style lang="scss" scoped></style>
