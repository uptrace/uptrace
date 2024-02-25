<template>
  <div>
    <portal to="navigation">
      <v-tabs :key="$route.fullPath" background-color="transparent">
        <v-tab :to="{ name: 'ProjectShow' }" exact-path>Settings</v-tab>
        <v-tab :to="{ name: 'ProjectDsn' }" exact-path>DSN</v-tab>
      </v-tabs>
    </portal>

    <v-container v-if="!project.data">
      <v-row>
        <v-col>
          <v-skeleton-loader type="card@3"></v-skeleton-loader>
        </v-col>
      </v-row>
    </v-container>
    <router-view v-else name="tab" :project="project.data" :dsn="project.dsn">
      <template #breadcrumbs>
        <v-breadcrumbs large :items="breadcrumbs" divider=">" class="px-1"></v-breadcrumbs>
      </template>
    </router-view>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useProject } from '@/org/use-projects'

export default defineComponent({
  name: 'Project',

  setup() {
    const project = useProject()
    useTitle(
      computed(() => {
        return project.data?.name ?? 'Project'
      }),
    )

    const breadcrumbs = computed(() => {
      const ss: any[] = [{ text: 'Projects' }]

      if (project.data) {
        ss.push({ text: project.data.name })
      } else {
        ss.push({ text: 'Project' })
      }

      return ss
    })

    return {
      breadcrumbs,

      project,
    }
  },
})
</script>

<style lang="scss" scoped>
.border-bottom {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
