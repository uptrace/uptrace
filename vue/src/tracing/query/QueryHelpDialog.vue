<template>
  <v-dialog v-model="dialog" :max-width="1280" @keydown.esc="dialog = false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" v-bind="attrs" v-on="on">
        <v-icon left>mdi-help-circle-outline</v-icon>
        <span>Help</span>
      </v-btn>
    </template>

    <v-card>
      <v-toolbar flat color="blue lighten-5">
        <v-toolbar-title>Spans Querying Cheat Sheet</v-toolbar-title>
        <v-btn
          href="https://uptrace.dev/get/querying-spans.html"
          target="_blank"
          class="ml-6 primary"
        >
          <span>Documentation</span>
          <v-icon right>mdi-open-in-new</v-icon>
        </v-btn>
        <v-spacer />
        <v-toolbar-items>
          <v-btn icon @click="dialog = false"><v-icon>mdi-close</v-icon></v-btn>
        </v-toolbar-items>
      </v-toolbar>

      <v-container fluid>
        <v-row>
          <v-col>
            <CheatsheetCard />
          </v-col>
        </v-row>

        <v-row>
          <v-spacer />
          <v-col cols="auto">
            <v-btn text color="primary" @click="dialog = false">Close</v-btn>
            <v-btn
              text
              color="primary"
              href="https://uptrace.dev/get/querying-spans.html"
              target="_blank"
            >
              <span>Read more</span>
              <v-icon right>mdi-open-in-new</v-icon>
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

// Components
import CheatsheetCard from '@/tracing/CheatsheetCard.vue'

export default defineComponent({
  name: 'QueryHelpDialog',
  components: { CheatsheetCard },

  setup() {
    const dialog = shallowRef(false)
    return { dialog }
  },
})
</script>

<style lang="scss" scoped></style>
