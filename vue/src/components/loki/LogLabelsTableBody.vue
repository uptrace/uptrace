<template>
  <tbody v-if="!labelKeys.length">
    <tr class="v-data-table__empty-wrapper">
      <td colspan="99">There are no attributes matching the filters.</td>
    </tr>
  </tbody>

  <tbody v-else>
    <tr v-for="key in labelKeys" :key="key">
      <th>{{ key }}</th>
      <td>{{ labels[key] }}</td>
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
