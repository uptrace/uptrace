<template>
  <div class="container--fixed-sm">
    <v-container class="mb-6">
      <v-row>
        <v-col class="d-flex justify-center text-h5"> ClickHouse query timeouts </v-col>
      </v-row>

      <v-row>
        <v-col class="d-flex justify-center">
          <v-icon size="56" color="red darken-1">mdi-clock-alert-outline</v-icon>
        </v-col>
      </v-row>

      <v-row>
        <v-col class="text-subtitle-1">
          Sorry, the ClickHouse query has timed out. The development team has been notified and is
          working on resolving this issue. To resolve it more efficiently, you can take advantage of
          our cloud-based version.
        </v-col>
      </v-row>

      <v-row>
        <v-col class="d-flex justify-center">
          <v-btn href="https://app.uptrace.dev/join" small class="primary">
            <v-icon left>mdi-cloud</v-icon>
            <span>Uptrace Cloud</span>
          </v-btn>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'ClickHouseTimeoutError',

  setup() {},
})
</script>

<style lang="scss" scoped></style>
