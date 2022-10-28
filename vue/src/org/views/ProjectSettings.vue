<template>
  <div class="container--fixed-sm">
    <PageToolbar>
      <v-toolbar-title>Project settings</v-toolbar-title>
    </PageToolbar>

    <v-container>
      <v-card v-if="project" flat>
        <v-form>
          <v-card-text>
            <v-text-field v-model="project.name" :disabled="disabled" label="Name" />

            <v-checkbox
              v-model="project.groupByEnv"
              label="Group spans by deployment.environment attribute"
              :disabled="disabled"
              hide-details="auto"
            >
            </v-checkbox>
            <v-checkbox
              v-model="project.groupFuncsByService"
              label="Group funcs spans by service.name attribute"
              :disabled="disabled"
              hide-details="auto"
            ></v-checkbox>
          </v-card-text>
        </v-form>
      </v-card>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, ref } from 'vue'

// Composables
import { useTitle } from '@vueuse/core'
import { useUser, Project } from '@/use/org'

export default defineComponent({
  name: 'ProjectSettings',

  setup() {
    const disabled = ref(true)
    const user = useUser()

    const project = computed((): Project | undefined => {
      return user.activeProject
    })

    const title = computed((): string => {
      return project.value?.name ?? 'Project'
    })

    useTitle(title)

    return { project, disabled }
  },
})
</script>

<style lang="scss" scoped>
.border {
  border-bottom: thin rgba(0, 0, 0, 0.12) solid;
}
</style>
