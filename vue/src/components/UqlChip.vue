<template>
  <v-tooltip :disabled="!tooltip" top>
    <template #activator="{ on, attrs }">
      <v-chip label v-bind="attrs" v-on="on" @click="apply">
        {{ text }}
      </v-chip>
    </template>
    {{ tooltip }}
  </v-tooltip>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseUql } from '@/use/uql'

// Misc
import { quote } from '@/util/string'

export default defineComponent({
  name: 'UqlChip',

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    column: {
      type: String,
      default: '',
    },
    group: {
      type: String,
      default: '',
    },
    op: {
      type: String,
      default: '',
    },
    value: {
      type: undefined,
      default: undefined,
    },
    query: {
      type: String,
      default: '',
    },
    tooltip: {
      type: String,
      default: undefined,
    },
  },

  setup(props, ctx) {
    const text = computed(() => {
      if (props.group) {
        return props.group
      }

      if (props.column) {
        const parts = [props.column, props.op]

        if (props.value) {
          parts.push(quote(props.value) as string)
        }

        return parts.join(' ')
      }

      if (props.query) {
        return props.query
      }

      return ''
    })

    function apply() {
      ctx.emit('click')

      if (props.group) {
        const editor = props.uql.createEditor()
        editor.resetGroupBy(props.group)
        props.uql.commitEdits(editor)
        return
      }

      if (props.column && !props.op) {
        return aggBy(props.column)
      }

      if (props.column && props.op) {
        const editor = props.uql.createEditor()
        editor.where(props.column, props.op, props.value)
        props.uql.commitEdits(editor)
        return
      }

      if (props.query) {
        props.uql.query = props.query
        return
      }
    }

    function aggBy(column: string) {
      const editor = props.uql.createEditor()
      editor.add(column)
      props.uql.commitEdits(editor)
    }

    return {
      text,

      apply,
    }
  },
})
</script>

<style lang="scss" scoped></style>
