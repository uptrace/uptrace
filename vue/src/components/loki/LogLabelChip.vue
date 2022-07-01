<template>
  <v-chip
    v-if="labelValue"
    v-model="chipSelected"
    small
    class="ma-1"
    :color="chipSelected ? 'blue' : 'grey lighten-4'"
    :class="{ active: chipSelected }"
    :pill="pill"
    :label="label"
    @click="setIsLabelSelected"
  >
    {{ labelValue }}
  </v-chip>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch } from '@vue/composition-api'

// Composables
import { Label } from '@/components/loki/logql'

export default defineComponent({
  name: 'LogLabelChip',
  props: {
    value: {
      type: Boolean,
      required: true,
    },
    labelValue: {
      type: String,
      required: true,
    },
    pill: {
      type: Boolean,
      default: false,
    },
    label: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const chipSelected = shallowRef(false)

    watch(
      () => props.value,
      (value) => {
        chipSelected.value = value
      },
      { immediate: true },
    )

    function setIsLabelSelected() {
      chipSelected.value = !chipSelected.value
      const label: Label = { name: '', value: props.labelValue, selected: chipSelected.value }
      ctx.emit('click:labelSelected', label)
      ctx.emit('input', chipSelected.value)
    }

    return { chipSelected, setIsLabelSelected }
  },
})
</script>
<style lang="scss" scoped>
.active {
  color: white;
}
</style>
