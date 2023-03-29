<template>
  <v-dialog v-model="dialog" :max-width="width" @keydown.esc="cancel">
    <v-card>
      <v-toolbar dark color="primary" dense flat>
        <v-toolbar-title class="white--text">
          {{ title }}
        </v-toolbar-title>
      </v-toolbar>

      <v-card-text class="pa-4">
        <span v-html="message"></span>
      </v-card-text>

      <v-card-actions class="pt-0">
        <v-spacer></v-spacer>
        <v-btn text data-cy="cancel-btn" @click="cancel">No</v-btn>
        <v-btn color="primary darken-1" text data-cy="yes-btn" @click="agree">Yes</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { useConfirm } from '@/use/confirm'

export default defineComponent({
  name: 'GlobalConfirm',

  setup() {
    return useConfirm()
  },
})
</script>

<style lang="scss" scoped></style>
