<template>
  <div class="mb-4">
    <div class="mb-1 text-subtitle-2 text--secondary font-weight-regular">
      <slot name="title">{{ title }}</slot>
    </div>

    <div>
      <slot />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'PanelSection',

  props: {
    title: {
      type: String,
      default: '',
    },
  },
})
</script>

<style lang="scss" scoped></style>
