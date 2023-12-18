<template>
  <div class="panel">
    <div
      class="py-2 d-flex align-center cursor-pointer"
      @click="expandedInternal = !expandedInternal"
    >
      <v-icon class="mr-1">{{ expandedInternal ? 'mdi-chevron-down' : 'mdi-chevron-up' }}</v-icon>
      <div class="text-subtitle-2 grey--text text--darken-4">{{ title }}</div>
    </div>
    <div v-if="expandedInternal" class="py-2 px-4">
      <slot />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

export default defineComponent({
  name: 'SinglePanel',

  props: {
    title: {
      type: String,
      required: true,
    },
    expanded: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const expandedInternal = shallowRef(props.expanded)

    return { expandedInternal }
  },
})
</script>

<style lang="scss" scoped>
.panel {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
