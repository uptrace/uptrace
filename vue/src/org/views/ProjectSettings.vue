<template>
  <div class="container--fixed-sm">
    <PageToolbar>
      <v-toolbar-title>Project settings</v-toolbar-title>
    </PageToolbar>

    <v-container fluid class="mb-6">
      <v-card v-if="project.data" flat>
        <v-card-text class="text-subtitle-1">
          <p>
            You can change project settings in the <code>uptrace.yml</code> config file. See
            <a href="https://uptrace.dev/get/config.html#managing-projects" target="_blank"
              >documentation</a
            >
            for details.
          </p>

          <v-form>
            <v-text-field v-model="project.data.name" :disabled="disabled" label="Name" filled />

            <v-select
              v-model="project.data.pinnedAttrs"
              label="Pinned attributes"
              :items="project.data.pinnedAttrs"
              multiple
              :disabled="disabled"
              filled
            />

            <v-checkbox
              v-model="project.data.groupByEnv"
              label="Group spans by deployment.environment attribute"
              :disabled="disabled"
              hide-details="auto"
            >
            </v-checkbox>
            <v-checkbox
              v-model="project.data.groupFuncsByService"
              label="Group funcs spans by service_name attribute"
              :disabled="disabled"
              hide-details="auto"
            ></v-checkbox>
          </v-form>
        </v-card-text>
      </v-card>
    </v-container>

    <PageToolbar id="dsn">
      <v-toolbar-title>Project DSN</v-toolbar-title>
    </PageToolbar>

    <v-container fluid class="mb-6">
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
import { defineComponent, computed } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useProject } from '@/org/use-projects'

export default defineComponent({
  name: 'ProjectSettings',

  props: {
    disabled: {
      type: Boolean,
      default: true,
    },
  },

  setup() {
    const project = useProject()

    const title = computed((): string => {
      return project.data?.name ?? 'Project'
    })
    useTitle(title)

    const headers = [{ text: 'DSN', value: 'dsn', sortable: false }]

    const tokens = computed(() => {
      if (!project.data) {
        return []
      }
      return [{ dsn: project.dsn }]
    })

    return {
      project,

      headers,
      tokens,
    }
  },
})
</script>

<style lang="scss" scoped>
.border {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
