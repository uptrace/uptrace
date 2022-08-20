<template>
  <v-btn small outlined @click="forceReload">Reload</v-btn>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { useForceReload } from '@/use/force-reload'

export default defineComponent({
  name: 'ForceReloadBtn',

  setup() {
    const { forceReload } = useForceReload()

    return { forceReload }
  },
})
</script>

<style lang="scss" scoped></style>
