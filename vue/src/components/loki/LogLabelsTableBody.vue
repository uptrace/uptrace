<template>
  <tbody v-if="!labelKeys.length">
    <tr class="v-data-table__empty-wrapper">
      <td colspan="99">There are no attributes matching the filters.</td>
    </tr>
  </tbody>

  <tbody v-else>
    <tr v-for="key in labelKeys" :key="key">
      <th>
        <span>{{ key }}</span>
      </th>
      <td>
        <span v-if="showFilters" class="mr-2">
          <v-btn
            icon
            small
            title="Filter for value"
            @click="$emit('click:filter', { key: key, op: '=', value: labels[key] })"
            ><v-icon>mdi-magnify-plus-outline</v-icon></v-btn
          >
          <v-btn
            icon
            small
            title="Filter out value"
            @click="$emit('click:filter', { key: key, op: '!=', value: labels[key] })"
            ><v-icon>mdi-magnify-minus-outline</v-icon></v-btn
          >
        </span>
        <span>{{ labels[key] }}</span>
      </td>
    </tr>
  </tbody>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from '@vue/composition-api'

export default defineComponent({
  name: 'LogLabelsTableBody',

  props: {
    labels: {
      type: Object as PropType<Record<string, string>>,
      required: true,
    },
    showFilters: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const labelKeys = computed((): string[] => {
      const keys = Object.keys(props.labels)
      keys.sort()
      return keys
    })

    return { labelKeys }
  },
})
</script>

<style lang="scss" scoped></style>
