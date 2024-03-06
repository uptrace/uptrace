<template>
  <v-menu v-if="numAchievCompleted < numAchievTotal" open-on-hover offset-x right :nudge-left="1">
    <template #activator="{ attrs, on }">
      <v-list-item title="Get started" class="px-2" v-bind="attrs" v-on="on">
        <v-list-item-avatar>
          <v-progress-circular
            :value="(numAchievCompleted / numAchievTotal) * 100"
          ></v-progress-circular>
        </v-list-item-avatar>
        <v-list-item-content>
          <v-list-item-title>
            Get started
            <span class="ml-2 text-body-2">{{ numAchievCompleted }} / {{ numAchievTotal }}</span>
          </v-list-item-title>
        </v-list-item-content>
        <v-list-item-icon>
          <v-icon dense>mdi-chevron-right</v-icon>
        </v-list-item-icon>
      </v-list-item>
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
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useAchievements } from '@/org/use-achievements'

// Misc
import { Project } from '@/org/use-projects'

export default defineComponent({
  name: 'GetStartedMenu',

  props: {
    project: {
      type: Object as PropType<Project>,
      required: true,
    },
  },

  setup(props) {
    const achievements = useAchievements(computed(() => props.project))

    const numAchievTotal = computed(() => {
      return achievements.items.length
    })

    const numAchievCompleted = computed(() => {
      return achievements.items.filter((achv) => {
        return achv.data && achievements.isCompleted(achv)
      }).length
    })

    return {
      achievements,

      numAchievTotal,
      numAchievCompleted,
    }
  },
})
</script>

<style lang="scss" scoped></style>
