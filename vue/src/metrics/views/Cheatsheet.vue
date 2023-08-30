<template>
  <div class="container--fixed-x-lg">
    <PageToolbar>
      <v-toolbar-title>Querying Metrics Cheatsheet</v-toolbar-title>
      <v-btn
        href="https://uptrace.dev/get/querying-metrics.html"
        target="_blank"
        class="ml-6 primary"
      >
        <span>Documentation</span>
        <v-icon right>mdi-open-in-new</v-icon>
      </v-btn>
    </PageToolbar>

    <v-container :fluid="$vuetify.breakpoint.mdAndDown">
      <v-row>
        <v-col>
          <CheatsheetCard />
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'

// Components
import CheatsheetCard from '@/metrics/CheatsheetCard.vue'

export default defineComponent({
  name: 'Cheatcheet',
  components: { CheatsheetCard },

  setup() {
    useTitle('Querying Metrics Cheatsheet')
    return {}
  },
})
</script>

<style lang="scss" scoped></style>
