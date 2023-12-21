<template>
  <v-autocomplete
    :value="value"
    :items="items"
    :search-input.sync="searchInput"
    filled
    dense
    hide-details="auto"
    @input="$emit('input', $event)"
  />
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed } from 'vue'

// Misc
import { Unit } from '@/util/fmt'

export default defineComponent({
  name: 'UnitSelect',

  props: {
    value: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const searchInput = shallowRef('')

    const stdItems = computed(() => {
      const items = [
        { text: 'none', value: Unit.None },
        { text: '', value: Unit.Bytes },
        { text: '', value: Unit.Nanoseconds },
        { text: '', value: Unit.Microseconds },
        { text: '', value: Unit.Milliseconds },
        { text: '', value: Unit.Seconds },
        { text: '', value: Unit.Utilization, hint: '0 <= n <= 1' },
        { text: '', value: Unit.Percents, hint: '0 <= n <= 100%' },
      ]

      for (let item of items) {
        if (!item.text) {
          item.text = item.value
        }
      }

      return items
    })

    const items = computed(() => {
      const items = stdItems.value.slice()

      if (props.value) {
        const i = items.findIndex((item) => item.value === props.value)
        if (i === -1) {
          items.push({ text: props.value, value: props.value as Unit })
        }
      }

      if (searchInput.value) {
        const i = items.findIndex((item) => item.value === searchInput.value)
        if (i === -1) {
          items.unshift({ text: searchInput.value, value: searchInput.value as Unit })
        }
      }

      return items
    })

    return { searchInput, items }
  },
})
</script>

<style lang="scss" scoped></style>
