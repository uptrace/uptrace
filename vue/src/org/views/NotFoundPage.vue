<template>
  <div class="container--fixed-sm">
    <PageToolbar>
      <v-toolbar-title>Page not found</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6">
      <v-card flat>
        <v-card-text class="text-subtitle-1">
          Sorry, the requested page can't be found. The development team has been notified.
        </v-card-text>
      </v-card>
    </v-container>

    <HelpLinks />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Components
import HelpLinks from '@/components/HelpLinks.vue'

export default defineComponent({
  name: 'NotFoundPage',
  components: { HelpLinks },

  setup() {},
})
</script>

<style lang="scss" scoped></style>
