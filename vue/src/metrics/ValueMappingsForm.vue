<template>
  <div>
    <v-row>
      <v-col>
        <ValueMappingRow
          v-for="(mapping, i) in mappings"
          :key="i"
          :value="mapping"
          @click:remove="removeMapping(i)"
        />
      </v-col>
    </v-row>

    <v-row align="center" justify="space-between">
      <v-col cols="auto">
        <v-btn @click="addMapping">
          <v-icon left>mdi-plus</v-icon>
          Add mapping
        </v-btn>
      </v-col>
      <v-col cols="auto">
        <v-btn text @click="$emit('click:close')">Cancel</v-btn>
        <v-btn color="primary" @click="emitMappings">Update</v-btn>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts">
import { clone } from 'lodash-es'
import { defineComponent, ref, watch, PropType } from 'vue'
import { eChart as colorScheme } from '@/util/colorscheme'

// Components
import ValueMappingRow from '@/metrics/ValueMappingRow.vue'

// Misc
import { emptyValueMapping, ValueMapping } from '@/metrics/types'

export default defineComponent({
  name: 'ValueMappingsForm',
  components: { ValueMappingRow },

  props: {
    value: {
      type: Array as PropType<ValueMapping[]>,
      required: true,
    },
  },

  setup(props, ctx) {
    const mappings = ref<ValueMapping[]>([])
    watch(
      () => props.value,
      (value) => {
        if (value.length) {
          mappings.value = clone(value)
        } else {
          mappings.value = [createValueMapping()]
        }
      },
      { immediate: true },
    )
    watch(
      () => mappings.value.length,
      (length) => {
        if (!length) {
          mappings.value = [createValueMapping()]
        }
      },
    )

    function addMapping() {
      mappings.value.push(createValueMapping())
    }

    function removeMapping(index: number) {
      mappings.value.splice(index, 1)
    }

    function createValueMapping() {
      const colorSet = new Set(colorScheme)
      for (let mapping of mappings.value) {
        if (mapping.color) {
          colorSet.delete(mapping.color)
        }
      }

      const mapping = emptyValueMapping()
      if (colorSet.size) {
        mapping.color = colorSet.values().next().value
      }
      return mapping
    }

    function emitMappings() {
      ctx.emit(
        'input',
        mappings.value.filter((m) => m.value || m.text),
      )
      ctx.emit('click:close')
    }

    return { mappings, addMapping, removeMapping, emitMappings }
  },
})
</script>

<style lang="scss" scoped></style>
