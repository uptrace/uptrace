<template>
  <v-hover v-slot="{ hover }">
    <v-card :flat="!hover" class="mx-auto">
      <v-card-text class="text-center py-1">
        <router-link v-if="to" :to="to">
          <v-img :src="icon" max-width="80" aspect-ratio="1" contain :alt="name" class="mx-auto" />
          <div class="mt-2 text-subtitle-1 text-no-wrap">{{ name }}</div>
        </router-link>
        <a v-else :href="href" :title="`${name} integration`" target="_blank">
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
  name: 'DevIcon',

  props: {
    name: {
      type: String,
      required: true,
    },
    icon: {
      type: String,
      required: true,
    },
    href: {
      type: String,
      default: '',
    },
    to: {
      type: Object,
      default: undefined,
    },
  },
})
</script>

<style lang="scss" scoped>
a {
  text-decoration: none;
}
</style>
