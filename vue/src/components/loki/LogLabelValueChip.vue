<template>
  <v-chip
    v-if="value"
    v-model="isValueSelected"
    pill
    small
    class="ma-1"
    :color="isValueSelected ? 'blue' : 'grey lighten-4'"
    :class="{ active: isValueSelected }"
    @click="setIsValueSelected"
    >{{ value }}</v-chip
  >
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, watch } from '@vue/composition-api'

export type LabelValue = {
  name: string
  selected: boolean
}
export default defineComponent({
  name: 'LogLabelValueChip',
  props: {
    value: {
      type: String,
      required: true,
    },
    selected: {
      type: Boolean,
      required: true,
    },
  },
  setup(props, ctx) {
    const isValueSelected = shallowRef(props?.selected)
    const setValueSelected = computed({
      get: () => isValueSelected.value,
      set: (value) => {
        isValueSelected.value = value
      },
    })

    watch(
      () => props.selected,
      (value) => {
        isValueSelected.value = value || false
      },
      { immediate: true },
    )
    function setIsValueSelected() {
      setValueSelected.value = !isValueSelected.value
      ctx.emit('click:valueSelected', {
        value: props.value || '',
        selected: setValueSelected.value,
      })
    }

    return { isValueSelected, setValueSelected, setIsValueSelected }
  },
})
</script>

<style lang="scss" scoped>
.active {
  color: white;
}
</style>
