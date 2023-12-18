<template>
  <v-btn :loading="forceReload.loading" :small="small" outlined @click="forceReload.do()"
    >Reload</v-btn
  >
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { injectForceReload } from '@/use/force-reload'

export default defineComponent({
  name: 'ForceReloadBtn',

  props: {
    small: {
      type: Boolean,
      default: false,
    },
  },

  setup() {
    const forceReload = injectForceReload()
    return { forceReload }
  },
})
</script>

<style lang="scss" scoped></style>
