<template>
  <v-chip
    v-model="isLabelSelected"
    label
    small
    class="ma-1"
    v-if="label"
    :color="isLabelSelected ? 'blue' : 'grey lighten-4'"
    :class="{ active: isLabelSelected }"
    @click="setIsLabelSelected"
  >
    {{ label.name }}
  </v-chip>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType, watch, computed } from '@vue/composition-api'

// Composables
import { UseDateRange } from '@/use/date-range'
import { Label } from '@/components/loki/logql'

export default defineComponent({
  name: 'LogLabelSelect',
  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    label: {
      type: Object as PropType<Label>,
      require: true,
    },
  },
  setup(props, ctx) {
    const isLabelSelected = shallowRef(props.label?.selected)
    const setLabelSelected = computed({
      get: () => isLabelSelected.value,
      set: (value) => {
        isLabelSelected.value = value
      },
    })
    watch(
      () => props.label?.selected,
      (label) => {
        isLabelSelected.value = label || false
      },
      { immediate: true },
    )

    function setIsLabelSelected() {
      setLabelSelected.value = !isLabelSelected.value
      const label: Label = { name: props?.label?.name || '', selected: setLabelSelected.value }
      ctx.emit('click:labelSelected', label)
    }

    return { isLabelSelected, setIsLabelSelected }
  },
})
</script>
<style lang="scss" scoped>
.active {
  color: white;
}
</style>
