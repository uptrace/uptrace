<template>
  <v-card class="d-flex flex-column pa-2 ma-1 mt-2" :elevation="1">
    <small class="text-caption font-weight-light mx-2">{{ label.name }}</small>
    <LogLabelValue
      v-for="(item, idx) in labelValues.selected"
      :key="idx"
      :value="item"
      @click:valueSelected="onClick(item.name)"
    />
  </v-card>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType, computed } from '@vue/composition-api'
import { UseDateRange } from '@/use/date-range'
import { useRouter } from '@/use/router'
import LogLabelValue from '@/components/loki/LogLabelValue.vue'
import { useLabelValues, Label } from '@/components/loki/logql'

// Composables

export default defineComponent({
  name: 'LogLabelValuesCont',
  components: { LogLabelValue },
  props: {
    label: {
      type: Object as PropType<Label>,
      required: true,
    },
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
  },
  setup(props, ctx) {
    const { route } = useRouter()
    const labelMenu = shallowRef(false)
    const isValueSelected = shallowRef(false)

    const setValueSelected = computed({
      get: () => isValueSelected.value,
      set: (value) => {
        isValueSelected.value = value
      },
    })

    const labelValues = useLabelValues(() => {
      const { projectId } = route.value.params
      return {
        url: `/${projectId}/loki/api/v1/label/${props.label.name}/values`,
        params: {
          ...props.dateRange.lokiParams(),
        },
      }
    })
    function addFilter(op: string, value: string) {
      setValueSelected.value = !isValueSelected.value
      ctx.emit('click', { op, value })
      labelMenu.value = false
    }

    function onClick(item: any) {
      addFilter('=', item)
    }

    return { onClick, labelValues, addFilter, isValueSelected }
  },
})
</script>
<style lang="scss" scoped></style>
