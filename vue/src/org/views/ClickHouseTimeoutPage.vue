<template>
  <div class="container--fixed-sm">
    <PageToolbar>
      <v-toolbar-title>ClickHouse query timeouts</v-toolbar-title>
    </PageToolbar>

    <v-container class="mb-6">
      <v-card flat>
        <v-card-text class="text-subtitle-1">
          Sorry, the ClickHouse query has timed out. The development team has been notified and is
          working on resolving this issue.
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
  name: 'ClickHouseTimeoutPage',
  components: { HelpLinks },

  setup() {},
})
</script>

<style lang="scss" scoped></style>
