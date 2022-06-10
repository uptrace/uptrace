<template>
  <v-chip
    v-if="label"
    v-model="isLabelSelected"
    label
    small
    class="ma-1"
    :color="isLabelSelected ? 'blue' : 'grey lighten-4'"
    :class="{ active: isLabelSelected }"
    @click="setIsLabelSelected"
  >
    {{ label }}
  </v-chip>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, computed } from '@vue/composition-api'

// Composables
import { Label } from '@/components/loki/logql'

export default defineComponent({
  name: 'LogLabelChip',
  props: {
    label: {
      type: String,
      default: '',
      require: true,
    },
    selected: {
      type: Boolean,
      required: true,
    },
  },
  setup(props, ctx) {
    const isLabelSelected = shallowRef(props?.selected)
    const setLabelSelected = computed({
      get: () => isLabelSelected.value,
      set: (value) => {
        isLabelSelected.value = value
      },
    })
    watch(
      () => props?.selected,
      (label) => {
        isLabelSelected.value = label || false
      },
      { immediate: true },
    )

    function setIsLabelSelected() {
      setLabelSelected.value = !isLabelSelected.value
      const label: Label = { name: props?.label || '', selected: setLabelSelected.value }
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
