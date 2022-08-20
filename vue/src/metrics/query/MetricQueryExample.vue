<template>
  <div class="mb-4">
    <p class="mb-1"><slot name="description" /></p>
    <XCode :code="query" />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'MetricQueryExample',

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
