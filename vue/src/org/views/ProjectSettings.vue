<template>
  <div class="container--fixed-md">
    <v-container fluid class="py-1">
      <slot name="breadcrumbs" />
    </v-container>

    <PageToolbar fluid>
      <v-toolbar-title>Project Settings</v-toolbar-title>
    </PageToolbar>

    <v-container fluid>
      <v-card flat>
        <v-card-text class="text-subtitle-1">
          <p>
            You can change project settings in the <code>uptrace.yml</code> config file. See
            <a href="https://uptrace.dev/get/config.html#managing-projects" target="_blank"
              >documentation</a
            >
            for details.
          </p>

          <v-form>
            <v-text-field v-model="project.name" :disabled="disabled" label="Name" filled />

            <v-select
              v-model="project.pinnedAttrs"
              label="Pinned attributes"
              :items="project.pinnedAttrs"
              multiple
              :disabled="disabled"
              filled
            />

            <v-checkbox
              v-model="project.groupByEnv"
              label="Group spans by deployment.environment attribute"
              :disabled="disabled"
              hide-details="auto"
            >
            </v-checkbox>
            <v-checkbox
              v-model="project.groupFuncsByService"
              label="Group funcs spans by service_name attribute"
              :disabled="disabled"
              hide-details="auto"
            ></v-checkbox>
          </v-form>
        </v-card-text>
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
  name: 'ProjectSettings',

  props: {
    project: {
      type: Object as PropType<Project>,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: true,
    },
  },

  setup(props) {
    const title = computed(() => {
      return `Settings | ${props.project.name}`
    })
    useTitle(title)
  },
})
</script>

<style lang="scss" scoped></style>
