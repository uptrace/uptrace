<template>
  <v-menu v-model="menu" offset-y>
    <template #activator="{ on, attrs }">
      <v-btn color="primary" dark v-bind="attrs" v-on="on">
        Create monitor
        <v-icon right>mdi-menu-down</v-icon>
      </v-btn>
    </template>
    <v-list>
      <v-list-item :to="{ name: 'MonitorMetricNew' }">
        <v-list-item-icon><v-icon>mdi-chart-line</v-icon></v-list-item-icon>
        <v-list-item-title>Create metrics monitor</v-list-item-title>
      </v-list-item>
      <v-list-item :to="{ name: 'MonitorErrorNew' }">
        <v-list-item-icon><v-icon>mdi-bug-outline</v-icon></v-list-item-icon>
        <v-list-item-title>Create errors monitor</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

export default defineComponent({
  name: 'MonitorNewMenu',

  setup() {
    const menu = shallowRef(false)
    return { menu }
  },
})
</script>

<style lang="scss" scoped></style>
