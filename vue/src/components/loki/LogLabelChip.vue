<template>
  <v-chip
    v-if="attrKey"
    v-model="chipSelected"
    small
    class="ma-1"
    :color="chipSelected ? 'blue' : 'grey lighten-4'"
    :class="{ active: chipSelected }"
    :pill="pill"
    :label="label"
    @click="setIsLabelSelected"
  >
    {{ attrKey }}
  </v-chip>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch } from '@vue/composition-api'

// Composables
import { Label } from '@/components/loki/logql'

export default defineComponent({
  name: 'LogLabelChip',
  props: {
    attrKey: {
      type: String,
      required: true,
    },
    selected: {
      type: Boolean,
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
    const chipSelected = shallowRef(props.selected)

    watch(
      () => props.selected,
      (value) => {
        chipSelected.value = value
      },
      { immediate: true },
    )

    function setIsLabelSelected() {
      chipSelected.value = !chipSelected.value
      const label: Label = { name: props.attrKey, selected: chipSelected.value }
      ctx.emit('click:labelSelected', label)
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
