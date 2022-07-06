<template>
  <router-link class="d-inline-block" v-bind="$attrs" v-on="$listeners">
    <v-img
      src="/uptrace/logo-small.svg"
      :max-width="width"
      :width="width"
      transition="scale-transition"
      alt="Uptrace"
      class="shrink"
    />
  </router-link>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'UptraceLogoSmall',

  props: {
    width: {
      type: [Number, String],
      default: 38,
    },
  },
})
</script>
