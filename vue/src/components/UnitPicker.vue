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
import { defineComponent, computed, PropType } from 'vue'

// Components
import BtnSelectMenu from '@/components/BtnSelectMenu.vue'

// Utilities
import { unitShortName, Unit } from '@/util/fmt'

export default defineComponent({
  name: 'UnitPicker',
  components: { BtnSelectMenu },

  props: {
    value: {
      type: String as PropType<Unit>,
      required: true,
    },
    base: {
      type: String as PropType<Unit>,
      default: undefined,
    },
    targetClass: {
      type: String,
      default: '',
    },
  },

  setup(props) {
    const availableItems = computed(() => {
      switch (props.base) {
        case undefined:
          return [
            { short: 'unit', text: 'none', value: Unit.None },
            { short: '', text: '', value: Unit.Bytes },
            { short: '', text: '', value: Unit.Nanoseconds },
            { short: '', text: '', value: Unit.Microseconds },
            { short: '', text: '', value: Unit.Milliseconds },
            { short: '', text: '', value: Unit.Seconds },
            { short: '', text: '', value: Unit.Percents },
          ]
        case Unit.Bytes:
          return [
            { short: '', text: '', value: Unit.Bytes },
            { short: '', text: '', value: Unit.Kilobytes },
            { short: '', text: '', value: Unit.Megabytes },
            { short: '', text: '', value: Unit.Gigabytes },
            { short: '', text: '', value: Unit.Terabytes },
          ]
        case Unit.Nanoseconds:
        case Unit.Microseconds:
        case Unit.Milliseconds:
        case Unit.Seconds:
          return [
            { short: '', text: '', value: Unit.Nanoseconds },
            { short: '', text: '', value: Unit.Microseconds },
            { short: '', text: '', value: Unit.Milliseconds },
            { short: '', text: '', value: Unit.Seconds },
          ]
        case Unit.Percents:
          return [{ short: '', text: '', value: Unit.Percents }]
        default:
          return [{ short: 'unit', text: 'none', value: Unit.None }]
      }
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
