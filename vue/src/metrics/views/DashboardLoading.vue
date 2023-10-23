<template>
  <v-container :fluid="$vuetify.breakpoint.lgAndDown">
    <v-skeleton-loader type="card,table"></v-skeleton-loader>
  </v-container>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'DashboardLoading',
})
</script>

<style lang="scss" scoped></style>
