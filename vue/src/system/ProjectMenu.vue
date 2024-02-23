<template>
  <v-menu open-on-hover offset-x right :nudge-left="1">
    <template #activator="{ attrs, on }">
      <v-list-item class="px-2" v-bind="attrs" v-on="on">
        <v-list-item-avatar>
          <v-avatar color="primary" size="36">
            <span class="white--text text-h5">{{ project.name.at(0) }}</span>
          </v-avatar>
        </v-list-item-avatar>

        <v-list-item-content>
          <v-list-item-title class="text-h6">{{ project.name }}</v-list-item-title>
        </v-list-item-content>

        <v-list-item-action>
          <v-icon>mdi-menu-down</v-icon>
        </v-list-item-action>
      </v-list-item>
    </template>

    <v-list>
      <v-list-item
        v-for="project in user.projects"
        :key="project.id"
        :to="{ name: 'Overview', params: { projectId: project.id } }"
        @click="saveLastProjectId(project.id)"
      >
        <v-list-item-content>
          <v-list-item-title>{{ project.name }}</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { useUser } from '@/org/use-users'
import { Project } from '@/org/use-projects'

export default defineComponent({
  name: 'ProjectMenu',

  props: {
    project: {
      type: Object as PropType<Project>,
      required: true,
    },
  },

  setup() {
    const user = useUser()

    function saveLastProjectId(projectId: number) {
      user.lastProjectId = projectId
    }

    return { user, saveLastProjectId }
  },
})
</script>

<style lang="scss" scoped></style>
