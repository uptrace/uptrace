<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" :disabled="disabled" v-bind="attrs" v-on="on">
        Group by
      </v-btn>
    </template>

    <SearchableList :items="attrKeys" @input="groupBy($event)"></SearchableList>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { UseUql } from '@/use/uql'

// Components
import SearchableList from '@/components/SearchableList.vue'

export default defineComponent({
  name: 'DashGroupByMenu',
  components: { SearchableList },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    attrKeys: {
      type: Array as PropType<string[]>,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const menu = shallowRef(false)

    function groupBy(attrKey: string) {
      const editor = props.uql.createEditor()
      editor.add(`group by ${attrKey}`)
      props.uql.commitEdits(editor)

      menu.value = false
    }

    return {
      menu,
      groupBy,
    }
  },
})
</script>

<style lang="scss" scoped></style>
