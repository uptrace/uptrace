<template>
  <v-navigation-drawer
    :value="value"
    app
    :temporary="temporary"
    :mini-variant="mini"
    @input="$emit('input', $event)"
  >
    <template v-if="mini">
      <v-row justify="center" align="center" no-gutters>
        <v-col cols="auto">
          <v-btn icon :title="mini ? 'Expand menu' : 'Minimize menu'" @click="mini = !mini">
            <v-icon>{{ mini ? 'mdi-unfold-more-vertical' : 'mdi-unfold-less-vertical' }}</v-icon>
          </v-btn>
        </v-col>
        <v-col cols="auto">
          <v-btn
            icon
            :title="temporary ? 'Keep menu open' : 'Undock menu'"
            @click="temporary = !temporary"
          >
            <v-icon>{{ temporary ? 'mdi-dock-left' : 'mdi-dock-window' }}</v-icon>
          </v-btn>
        </v-col>
      </v-row>
      <v-divider />
    </template>

    <v-system-bar v-else window color="grey lighten-4">
      <v-spacer />
      <v-btn icon :title="mini ? 'Expand menu' : 'Minimize menu'" @click="mini = !mini">
        <v-icon>{{ mini ? 'mdi-unfold-more-vertical' : 'mdi-unfold-less-vertical' }}</v-icon>
      </v-btn>
      <v-btn
        icon
        :title="temporary ? 'Keep menu open' : 'Hide menu'"
        @click="temporary = !temporary"
      >
        <v-icon>{{ temporary ? 'mdi-dock-left' : 'mdi-dock-window' }}</v-icon>
      </v-btn>
    </v-system-bar>

    <v-list class="list--hoverable">
      <v-list-item v-if="user.isAuth && !user.projects.length" :to="{ name: 'ProjectNew' }">
        <v-list-item-action>
          <v-icon>mdi-plus</v-icon>
        </v-list-item-action>
        <v-list-item-content>
          <v-list-item-title class="text-h6">New project</v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <ProjectMenu v-if="project.data" :project="project.data" />
    </v-list>
    <v-divider />

    <ProjectNavigationList v-if="project.data" :project="project.data" class="list--hoverable" />

    <template #append>
      <v-divider />
      <v-list class="list--hoverable">
        <HowToMenu v-if="project.data" :project="project.data" />
        <GetStartedMenu v-if="project.data" :project="project.data" />

        <UserMenu v-if="user.isAuth" />
      </v-list>
    </template>
  </v-navigation-drawer>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

// Composables
import { useStorage } from '@/use/local-storage'
import { useUser } from '@/org/use-users'
import { useProject } from '@/org/use-projects'

// Components
import ProjectMenu from '@/system/ProjectMenu.vue'
import ProjectNavigationList from '@/system/ProjectNavigationList.vue'
import GetStartedMenu from '@/system/GetStartedMenu.vue'
import HowToMenu from '@/system/HowToMenu.vue'
import UserMenu from '@/system/UserMenu.vue'

export default defineComponent({
  name: 'AppNavigation',
  components: { ProjectMenu, ProjectNavigationList, GetStartedMenu, HowToMenu, UserMenu },

  props: {
    value: {
      type: Boolean,
      required: true,
    },
  },

  setup() {
    const temporary = useStorage('navigation-temporary', false)
    const mini = useStorage('navigation-mini', false)

    const user = useUser()
    const project = useProject()

    return {
      temporary,
      mini,

      user,
      project,
    }
  },
})
</script>

<style lang="scss">
.list--hoverable .v-list-item:hover {
  background-color: rgba(0, 0, 0, 0.1);
}
</style>
