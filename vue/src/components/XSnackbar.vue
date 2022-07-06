<template>
  <v-snackbar v-model="snackbar.active" :color="snackbar.color" :timeout="snackbar.timeout" top>
    <span>{{ snackbar.notification }}</span>

    <template #action="{ attrs }">
      <v-btn text v-bind="attrs" @click="snackbar.active = false">Close</v-btn>
    </template>
  </v-snackbar>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { useSnackbar } from '@/use/snackbar'

export default defineComponent({
  name: 'XSnackbar',

  setup() {
    return { snackbar: useSnackbar() }
  },
})
</script>

<style lang="scss" scoped></style>
