<template>
  <v-hover v-slot="{ hover }">
    <v-card :flat="!hover" :color="hover ? 'white' : 'transparent'" class="mx-auto">
      <v-card-text class="text-center">
        <a :href="to" :title="`Uptrace for ${name}`" target="_blank">
          <v-img :src="icon" max-width="80" aspect-ratio="1" contain :alt="name" class="mx-auto" />
          <div class="mt-2 text-subtitle-1 text-no-wrap">{{ name }}</div>
        </a>
      </v-card-text>
    </v-card>
  </v-hover>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'Devicon',

  props: {
    name: {
      type: String,
      required: true,
    },
    icon: {
      type: String,
      required: true,
    },
    to: {
      type: String,
      required: true,
    },
  },
})
</script>

<style lang="scss" scoped>
a {
  text-decoration: none;
}
</style>
