<template>
  <BtnSelectMenu
    :value="value"
    :items="items"
    color="white"
    :target-class="targetClass"
    @input="$emit('input', $event)"
  />
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Components
import BtnSelectMenu from '@/components/BtnSelectMenu.vue'

// Misc
import { unitShortName, Unit } from '@/util/fmt'

export default defineComponent({
  name: 'UnitPicker',
  components: { BtnSelectMenu },

  props: {
    value: {
      type: String,
      required: true,
    },
    targetClass: {
      type: String,
      default: '',
    },
  },

  setup(props) {
    const availableItems = computed(() => {
      const items = [
        { short: 'unit', text: 'none', value: Unit.None },
        { short: '', text: '', value: Unit.Bytes },
        { short: '', text: '', value: Unit.Nanoseconds },
        { short: '', text: '', value: Unit.Microseconds },
        { short: '', text: '', value: Unit.Milliseconds },
        { short: '', text: '', value: Unit.Seconds },
        { short: '', text: '', value: Unit.Utilization, hint: '0 <= n <= 1' },
        { short: '', text: '', value: Unit.Percents, hint: '0 <= n <= 100%' },
      ]
      if (props.value) {
        const i = items.findIndex((item) => item.value === props.value)
        if (i === -1) {
          items.push({ short: '', text: '', value: props.value as Unit })
        }
      }
      return items
    })

    const items = computed(() => {
      return availableItems.value.map((item) => {
        if (!item.short) {
          item.short = unitShortName(item.value)
        }
        if (!item.text) {
          item.text = item.value
        }
        return item
      })
    })

    return { items }
  },
})
</script>

<style lang="scss" scoped></style>
