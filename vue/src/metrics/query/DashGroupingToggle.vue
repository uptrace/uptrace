<template>
  <v-chip-group v-model="grouping" active-class="primary--text" center-active>
    <v-chip :value="null" class="my-0" color="grey lighten-4">No grouping</v-chip>
    <v-chip
      v-for="item in attrKeys"
      :key="item.value"
      :value="item.value"
      color="my-0 grey lighten-4"
    >
      {{ item.text }}
    </v-chip>
  </v-chip-group>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { Item } from '@/use/datasource'
import { UseUql } from '@/use/uql'
import { ColumnInfo } from '@/metrics/types'

export default defineComponent({
  name: 'DashGroupingToggle',

  props: {
    attrKeys: {
      type: Array as PropType<Item[]>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    columns: {
      type: Array as PropType<ColumnInfo[]>,
      required: true,
    },
  },

  setup(props) {
    const grouping = computed({
      get() {
        const grouping = props.columns.filter((col) => col.isGroup)
        switch (grouping.length) {
          case 0:
            return null
          case 1:
            return grouping[0].name
          default:
            return undefined
        }
      },
      set(grouping) {
        if (grouping === undefined) {
          return
        }

        const editor = props.uql.createEditor()

        if (grouping) {
          editor.resetGroupBy(grouping)
        } else {
          editor.resetGroupBy()
        }

        props.uql.commitEdits(editor)
      },
    })

    return { grouping }
  },
})
</script>

<style lang="scss" scoped></style>
