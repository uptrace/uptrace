<template>
  <div class="container--fixed-md">
    <v-container fluid class="py-1">
      <slot name="breadcrumbs" />
    </v-container>

    <PageToolbar fluid>
      <v-toolbar-title>Data Source Name</v-toolbar-title>
    </PageToolbar>

    <v-container fluid>
      <v-card flat>
        <v-card-text class="text-subtitle-1">
          Use the DSN (Data Source Name) below to
          <router-link :to="{ name: 'TracingHelp' }">configure your app</router-link>
          to send data to Uptrace. You can change the token in the <code>uptrace.yml</code> config
          file.
        </v-card-text>

        <v-data-table hide-default-footer :headers="headers" :items="tokens"> </v-data-table>
      </v-card>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { Project } from '@/org/use-projects'

export default defineComponent({
  name: 'ProjectDsn',

  props: {
    project: {
      type: Object as PropType<Project>,
      required: true,
    },
    dsn: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    useTitle(
      computed((): string => {
        return `Data Source Name | ${props.project.name}`
      }),
    )

    const headers = [{ text: 'DSN', value: 'dsn', sortable: false }]

    const tokens = computed(() => {
      if (!props.project) {
        return []
      }
      return [{ dsn: props.dsn }]
    })

    return { headers, tokens }
  },
})
</script>

<style lang="scss" scoped></style>
