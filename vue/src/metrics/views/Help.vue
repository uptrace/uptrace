<template>
  <HelpCard />
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Components
import HelpCard from '@/metrics/HelpCard.vue'

export default defineComponent({
  name: 'MetricsHelp',
  components: { HelpCard },
})
</script>

<style lang="scss" scoped></style>
