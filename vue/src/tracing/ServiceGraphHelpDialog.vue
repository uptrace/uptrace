<template>
  <v-dialog v-model="dialog" width="700">
    <template #activator="{ on, attrs }">
      <v-btn block v-bind="attrs" v-on="on">
        <v-icon left>mdi-help-circle-outline</v-icon>
        Help
      </v-btn>
    </template>

    <v-card>
      <v-toolbar color="bg--primary" flat>
        <v-toolbar-title>Service Graph</v-toolbar-title>

        <v-spacer />

        <v-toolbar-items>
          <v-btn icon @click="dialog = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-toolbar-items>
      </v-toolbar>

      <v-container fluid>
        <v-row>
          <v-col>
            <ServiceGraphHelpCard />
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from 'vue'

// Components
import ServiceGraphHelpCard from '@/tracing/ServiceGraphHelpCard.vue'

export default defineComponent({
  name: 'ServiceGraphHelpDialog',
  components: { ServiceGraphHelpCard },

  setup() {
    const dialog = shallowRef(false)

    return { dialog }
  },
})
</script>

<style lang="scss" scoped>
li {
  margin-bottom: 8px !important;
}
</style>
