<template>
  <v-card class="d-flex flex-column pa-2 ma-1 mt-2" :elevation="1">
    <small>{{ label.label }}</small>
    <!-- <v-chip
      label
      x-small
      class="ma-1"
      v-for="(item, idx) in labelValues.items"
      :key="idx"
      @click="onClick(item)"
      >{{ item }}</v-chip
    > -->
    <LogLabelValue
      v-for="(item, idx) in labelValues.selected"
      :key="idx"
      :value="item"
      @click:valueSelected="onClick(item.name)"
    />
  </v-card>
</template>

<script lang="ts">
// import {truncate } from 'lodash'
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
    // make call from component as in label
    const { route } = useRouter()
    const labelMenu = shallowRef(false)
    const isValueSelected = shallowRef(false)

    const setValueSelected = computed({
      get: () => isValueSelected.value,
      set: (value) => {
        isValueSelected.value = value
      },
    })

    //  add label values unit as a split component
    // pass this values as props down to that split component

    const labelValues = useLabelValues(() => {
      const { projectId } = route.value.params
      return {
        url: `/${projectId}/loki/api/v1/label/${props.label.label}/values`,
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
      console.log(item)
      addFilter('=', item)

      console.log(item, props, ctx)
    }

    return { onClick, labelValues, addFilter, isValueSelected }
  },
})
</script>
<style lang="scss" scoped>
.active {
  background: #1e88e5 !important;
  color: white;
}
</style>
