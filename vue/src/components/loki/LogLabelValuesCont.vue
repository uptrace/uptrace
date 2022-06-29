<template>
  <v-card class="d-flex flex-column pa-2 ma-1 mt-2" :elevation="1">
    <small class="text-caption font-weight-light mx-2">{{ label }}</small>
    <LogLabelChip
      v-for="(item, idx) in labels"
      :key="idx"
      v-model="item.selected"
      :attr-key="item.name"
      pill
      @click:labelSelected="onClick(item)"
    />
  </v-card>
</template>

<script lang="ts">
import {
  defineComponent,
  shallowRef,
  computed,
  reactive,
  PropType,
  watch,
} from '@vue/composition-api'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useRouter } from '@/use/router'
import { useLabels, Label } from '@/components/loki/logql'

// Components
import LogLabelChip from '@/components/loki/LogLabelChip.vue'
export interface LabelItem {
  name: string
  selected: boolean
  label: string
}

export enum Operators {
  equals = '=',
  matches = '=~',
}

export default defineComponent({
  name: 'LogLabelValuesCont',
  components: { LogLabelChip },

  props: {
    label: {
      type: String,
      required: true,
    },
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    query: {
      type: String,
      required: true,
    },
  },

  setup(props, ctx) {
    const { route } = useRouter()
    const labelMenu = shallowRef(false)

    const labelValues = useLabels(() => {
      const { projectId } = route.value.params
      return {
        url: `/${projectId}/loki/api/v1/label/${props.label}/values`,
        params: {
          ...props.dateRange.lokiParams(),
        },
      }
    })

    const internalLabels = computed((): LabelItem[] => {
      return labelValues.items.map((value: string): LabelItem => {
        return { name: value, selected: false, label: props.label || '' }
      })
    })

    const labels = computed((): Label[] => {
      return internalLabels.value.map((label) => reactive(label))
    })
    watch(
      () => props.query,
      (query) => {
        // parse the query in 're'
        labels.value.forEach((label) => {
          if (query?.includes(label.name) && query?.includes(label?.label)) {
            label.selected = true
          } else {
            label.selected = false
          }
        })
      },
      { immediate: true },
    )

    function addFilter(op: string, value: string, selected: boolean, labelValues: any) {
      labels.value.forEach((value) => {
        if (props.query?.includes(value.name)) {
          value.selected = true
        }
      })

      ctx.emit('click', { op, value, selected, labelValues, label: props.label })
      labelMenu.value = false
    }

    function onClick(item: any) {
      const { name, selected } = item
      addFilter('=', name, selected, labels)
    }

    return { labels, onClick, addFilter }
  },
})
</script>

<style lang="scss" scoped></style>
