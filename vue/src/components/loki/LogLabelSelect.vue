<template>
  <v-chip
    v-model="isLabelSelected"
    label
    x-small
    class="ma-1"
    v-if="label"
    @click="setIsLabelSelected"
    :class="{ active: isLabelSelected }"
  >
    {{ label.name }}
  </v-chip>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType, watch, computed } from '@vue/composition-api'

// Composables
import { UseDateRange } from '@/use/date-range'
import { LabelSelection } from '@/components/loki/logql'

// here should be the complete template with values list incluided
export default defineComponent({
  name: 'LogLabelSelect',
  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    label: {
      type: Object as PropType<LabelSelection>,
      require: true,
    },
  },
  setup(props, ctx) {
    const isLabelSelected = shallowRef(props.label?.selected)
    // const labelValuesSelected = shallowRef([])

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

      ctx.emit('click:labelSelected', {
        label: props?.label?.name || '',
        selected: setLabelSelected.value,
      })
    }

    return { isLabelSelected, setIsLabelSelected }
  },
})
</script>
<style lang="scss" scoped>
.active {
  background: #1e88e5 !important;
  color: white;
}
</style>
