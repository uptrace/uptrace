<template>
  <v-tooltip bottom>
    <template #activator="{ on, attrs }">
      <v-icon v-bind="attrs" v-on="on">mdi-lock-open-outline</v-icon>
    </template>
    <div>
      This dashboard was automatically created from a template and can't be changed.<br />
      But you can clone this dashboard and then change it as you want.
    </div>
  </v-tooltip>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'DashLockedIcon',
})
</script>

<style lang="scss" scoped></style>
