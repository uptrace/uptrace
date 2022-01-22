<template>
  <div>
    <v-simple-table class="v-data-table--border table-scroll-target">
      <colgroup>
        <col />
        <col />
      </colgroup>

      <thead class="v-data-table-header">
        <tr>
          <th>Key</th>
          <th class="target">Value</th>
        </tr>
      </thead>

      <tbody v-if="!attrKeys.length">
        <tr class="v-data-table__empty-wrapper">
          <td colspan="99">There are no attributes matching the filters.</td>
        </tr>
      </tbody>

      <tbody>
        <tr v-for="key in attrKeys" :key="key">
          <th>{{ key }}</th>
          <td><XText :name="key" :value="attrs[key]" /></td>
        </tr>
      </tbody>
    </v-simple-table>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from '@vue/composition-api'

// Utitlies
import { xkey } from '@/models/otelattr'
import { AttrMap } from '@/models/span'

export default defineComponent({
  name: 'AttrTable',

  props: {
    attrs: {
      type: Object as PropType<AttrMap>,
      required: true,
    },
  },

  setup(props) {
    const attrKeys = computed((): string[] => {
      const keys = Object.keys(props.attrs)
      keys.sort()
      return keys
    })

    return {
      xkey,
      attrKeys,
    }
  },
})
</script>

<style lang="scss" scoped></style>
