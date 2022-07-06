<template>
  <div v-frag>
    <v-toolbar-title
      v-if="!autocompleteActive && user.projects.length"
      class="cursor-pointer"
      @click="onClickActiveProject"
    >
      <span>{{ activeProject.name }}</span>
      <v-icon>mdi-menu-down</v-icon>
    </v-toolbar-title>

    <v-autocomplete
      v-if="autocompleteActive && user.projects.length"
      ref="el"
      v-model="activeProject"
      :items="user.projects"
      item-text="name"
      item-value="id"
      return-object
      :search-input.sync="searchInput"
      placeholder="Select a project"
      hide-details
      filled
      style="max-width: 300px"
      class="v-input--dense"
      @input="onChange"
      @blur="onBlur"
    >
    </v-autocomplete>
  </div>
</template>

<script lang="ts">
import Vue from 'vue'
import { defineComponent, ref, computed } from 'vue'

// Composables
import { useRouter } from '@/use/router'
import { useUser, Project } from '@/use/org'

interface Autocomplete {
  focus: () => void
  blur: () => void
  activateMenu: () => void
}

export default defineComponent({
  name: 'ProjectPicker',

  setup() {
    const { router, route } = useRouter()
    const user = useUser()

    const el = ref<Autocomplete>()
    const searchInput = ref('')
    const selectingProject = ref(false)

    const activeProject = computed({
      get(): Project | undefined {
        const projectId = parseInt(route.value.params.projectId)
        if (!projectId) {
          return
        }

        for (let p of user.projects) {
          if (p.id === projectId) {
            return p
          }
        }
      },
      set(project: Project) {
        let routeName = route.value.name as string

        const paramKeys = Object.keys(route.value.params)
        if (paramKeys.length !== 1 || paramKeys[0] !== 'projectId') {
          routeName = 'Overview'
        }

        router
          .push({
            name: routeName,
            params: { projectId: project.id.toString() },
          })
          .catch(() => {})
      },
    } as any)

    const autocompleteActive = computed(() => {
      return !activeProject.value || selectingProject.value
    })

    function onBlur() {
      setTimeout(() => {
        selectingProject.value = false
      }, 200)
    }

    function onChange() {
      Vue.nextTick(() => {
        if (el.value) {
          el.value.blur()
        }
      })
    }

    function onClickActiveProject() {
      selectingProject.value = true

      Vue.nextTick(() => {
        if (el.value) {
          el.value.focus()
          el.value.activateMenu()
        }
      })
    }

    return {
      route,

      el,
      user,
      activeProject,

      searchInput,
      autocompleteActive,
      onBlur,
      onChange,
      onClickActiveProject,
    }
  },
})
</script>

<style lang="scss" scoped></style>
