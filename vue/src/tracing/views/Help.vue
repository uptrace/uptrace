<template>
  <HelpCard />
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'

// Components
import HelpCard from '@/tracing/HelpCard.vue'

export default defineComponent({
  name: 'Tracing',
  components: { HelpCard },

  setup() {
    useTitle('Getting started with Uptrace')
  },
})
</script>

<style lang="scss" scoped></style>
