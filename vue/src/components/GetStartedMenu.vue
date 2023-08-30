<template>
  <v-menu v-model="menu" offset-y dark>
    <template #activator="{ on }">
      <v-btn small outlined tile color="grey lighten-2" v-on="on">
        <span>Get started</span>
        <v-icon right color="grey lighten-2">mdi-menu-down</v-icon>
      </v-btn>
    </template>

    <v-sheet max-width="550">
      <v-list>
        <v-list-item
          v-for="item in achievements.items"
          :key="item.name"
          :class="{ completed: achievements.isCompleted(item) }"
          v-bind="item.attrs"
        >
          <v-list-item-icon>
            <v-icon v-if="achievements.isCompleted(item)" size="30" color="green lighten-2">
              mdi-check-circle
            </v-icon>
            <v-icon v-else size="30" color="grey lighten-1">
              mdi-checkbox-blank-circle-outline
            </v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>{{ item.title }}</v-list-item-title>
            <v-list-item-subtitle>{{ item.subtitle }}</v-list-item-subtitle>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-sheet>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

// Composables
import { Project } from '@/org/use-projects'
import { UseAchievements } from '@/org/use-achievements'

export default defineComponent({
  name: 'GetStartedMenu',

  props: {
    project: {
      type: Object as PropType<Project>,
      default: undefined,
    },
    achievements: {
      type: Object as PropType<UseAchievements>,
      required: true,
    },
  },

  setup(props) {
    const menu = shallowRef(false)

    return {
      menu,
    }
  },
})
</script>

<style lang="scss" scoped>
.v-list-item__title {
  color: map-get($grey, 'lighten-3');
}

.v-list-item.completed {
  & ::v-deep .v-list-item__title {
    color: map-get($grey, 'lighten-1');
    text-decoration: line-through;
  }
}

.v-avatar {
  color: rgba(255, 255, 255, 0.95);
}
</style>
