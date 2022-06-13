<template>
  <v-card class="d-flex flex-column pa-2 ma-1 mt-2" :elevation="1">
    <small class="text-caption font-weight-light mx-2">{{ label }}</small>
    <LogLabelChip
      v-for="(item, idx) in labels"
      :key="idx"
      v-model="item.selected"
      :attr-key="item.name"
      pill
      @click:labelSelected="onClick(item.name)"
    />
  </v-card>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, reactive, PropType } from '@vue/composition-api'

// Composables
import { UseDateRange } from '@/use/date-range'
import { useRouter } from '@/use/router'
import { useLabels, Label } from '@/components/loki/logql'

// Components
import LogLabelChip from '@/components/loki/LogLabelChip.vue'

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

    const internalLabels = computed((): Label[] => {
      return labelValues.items.map((value: string): Label => {
        return { name: value, selected: false }
      })
    })

    const labels = computed((): Label[] => {
      return internalLabels.value.map((label) => reactive(label))
    })

    function addFilter(op: string, value: string) {
      ctx.emit('click', { op, value })
      labelMenu.value = false
    }

    function onClick(item: any) {
      addFilter('=', item)
    }

    return { labels, onClick, addFilter }
  },
})
</script>

<style lang="scss" scoped></style>
