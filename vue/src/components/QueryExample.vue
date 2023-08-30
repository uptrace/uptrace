<template>
  <div class="mb-5">
    <p class="mb-1"><slot name="description" /></p>
    <PrismCode :code="query" class="mt-1" />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'SpanQueryExample',

  props: {
    query: {
      type: String,
      required: true,
    },
  },
})
</script>

<style lang="scss" scoped>
.v-sheet {
  background-color: map-get($grey, 'lighten-4') !important;
}
</style>
