<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" v-bind="attrs" v-on="on">
        <span>{{ label }}</span>
        <v-icon right class="ml-0">mdi-menu-down</v-icon>
      </v-btn>
    </template>

    <XList
      :loading="suggestions.loading"
      :items="suggestions.filteredItems"
      :num-item="suggestions.items.length"
      :search-input.sync="suggestions.searchInput"
      return-object
      @input="addFilter($event.text, '=')"
    >
      <template #item="{ item }">
        <v-list-item-content>
          <v-list-item-title>
            {{ truncate(item.text, { length: 60 }) }}
          </v-list-item-title>
        </v-list-item-content>

        <v-list-item-action class="my-0" @click.stop="addFilter(item.text, '!=')">
          <v-btn icon>
            <v-icon small>mdi-not-equal</v-icon>
          </v-btn>
        </v-list-item-action>
      </template>
    </XList>
  </v-menu>
</template>

<script lang="ts">
import { truncate } from 'lodash'
import { defineComponent, shallowRef, PropType } from '@vue/composition-api'

// Composables
import { AxiosParams } from '@/use/axios'
import { UseUql } from '@/use/uql'
import { useSuggestions } from '@/use/suggestions'

// Components
import XList from '@/components/XList.vue'

export default defineComponent({
  name: 'AttrFilterMenu',
  components: { XList },

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
  },

  setup(props, ctx) {
    const menu = shallowRef(false)

    const suggestions = useSuggestions(() => {
      if (!menu.value) {
        return null
      }

      return {
        url: `/api/tracing/suggestions/values`,
        params: {
          ...props.axiosParams,
          column: props.attrKey,
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

    return { menu, suggestions, addFilter, truncate }
  },
})
</script>

<style lang="scss" scoped></style>
