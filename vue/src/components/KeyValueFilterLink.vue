<template>
  <span>
    <KeyValueFilterLinkItem
      v-for="val in values"
      :key="val"
      :date-range="dateRange"
      :name="name"
      :value="val"
      :project-id="projectId"
      :system="system"
      :group-id="groupId"
      :is-event="isEvent"
      :filterable="filterable"
    />
  </span>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { UseDateRange } from '@/use/date-range'

// Components
import KeyValueFilterLinkItem from '@/components/KeyValueFilterLinkItem.vue'

export default defineComponent({
  name: 'KeyValueFilterLink',
  components: { KeyValueFilterLinkItem },

  props: {
    dateRange: {
      type: Object as PropType<UseDateRange>,
      required: true,
    },
    projectId: {
      type: [Number, String],
      required: true,
    },
    system: {
      type: String,
      required: true,
    },
    groupId: {
      type: String,
      default: undefined,
    },
    name: {
      type: String,
      required: true,
    },
    value: {
      type: undefined,
      required: true,
    },
    isEvent: {
      type: Boolean,
      required: true,
    },
    filterable: {
      type: Boolean,
      default: true,
    },
  },

  setup(props) {
    const values = computed(() => {
      if (Array.isArray(props.value)) {
        return props.value
      }
      return [props.value]
    })

    return { values }
  },
})
</script>

<style lang="scss" scoped></style>
