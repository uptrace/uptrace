<template>
  <v-chip
    v-model="isValueSelected"
    v-if="value"
    label
    x-small
    class="ma-1"
    :class="{ active: isValueSelected }"
    @click="setIsValueSelected"
    >{{ value.name }}</v-chip
  >
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType, computed, watch } from '@vue/composition-api'

export type LabelValue = {
  name: string
  selected: boolean
}
export default defineComponent({
  name: 'LogLabelValue',
  props: {
    value: {
      type: Object as PropType<LabelValue>,
      required: true,
    },
  },
  setup(props, ctx) {
    const isValueSelected = shallowRef(props.value.selected)
    const setValueSelected = computed({
      get: () => isValueSelected.value,
      set: (value) => {
        isValueSelected.value = value
      },
    })

    watch(
      () => props.value.selected,
      (value) => {
        isValueSelected.value = value || false
      },
      { immediate: true },
    )
    function setIsValueSelected() {
      setValueSelected.value = !isValueSelected.value
      ctx.emit('click:valueSelected', {
        value: props.value.name || '',
        selected: setValueSelected.value,
      })
    }

    function onClick(item: string) {
      console.log(item)
    }
    return { isValueSelected, setValueSelected, onClick, setIsValueSelected }
  },
})
</script>

<style lang="scss" scoped>
.active {
  background: #1e88e5 !important;
  color: white;
}
</style>
