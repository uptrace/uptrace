<template>
  <v-system-bar height="40" app absolute dark class="grey--text text--lighten-2">
    <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="py-0 fill-height">
      <div class="mr-8">
        <span v-if="user.name"
          >Welcome, <strong>{{ user.name }}!</strong></span
        >
        <span v-else>Welcome!</span>
      </div>

      <div>
        <GetStartedMenu :project="project" :achievements="achievements" />
      </div>

      <div class="ml-4">
        <span>Completed </span>
        <strong>{{ achievements.completed.length }} of {{ achievements.items.length }}</strong>
        <span> tasks</span>
        <v-progress-linear
          :value="(achievements.completed.length / achievements.items.length) * 100"
          color="blue lighten-1"
          background-opacity="0.5"
        ></v-progress-linear>
      </div>

      <div class="ml-10">
        <v-menu offset-y dark>
          <template #activator="{ on }">
            <v-btn small outlined tile color="grey lighten-2" v-on="on">
              <span>How to?</span>
              <v-icon right color="grey lighten-2">mdi-menu-down</v-icon>
            </v-btn>
          </template>

          <v-sheet max-width="550">
            <v-list>
              <v-list-item :to="{ name: 'ProjectShow', hash: '#dsn' }">
                <v-list-item-content>
                  <v-list-item-title>How to find my project DSN?</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item :to="{ name: 'TracingHelp' }">
                <v-list-item-content>
                  <v-list-item-title>How to setup tracing?</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item :to="{ name: 'MetricsHelp' }">
                <v-list-item-content>
                  <v-list-item-title>How to collect metrics?</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item href="https://uptrace.dev/get/logging.html" target="_blank">
                <v-list-item-content>
                  <v-list-item-title>How to monitor logs?</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item :to="{ name: 'TracingCheatsheet' }">
                <v-list-item-content>
                  <v-list-item-title>How to query spans?</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item :to="{ name: 'MetricsCheatsheet' }">
                <v-list-item-content>
                  <v-list-item-title>How to query metrics?</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-sheet>
        </v-menu>
      </div>

      <v-spacer />

      <v-btn href="https://app.uptrace.dev/join" small class="primary">
        <v-icon left>mdi-cloud</v-icon>
        <span>Uptrace Cloud</span>
      </v-btn>
    </v-container>
  </v-system-bar>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

// Composables
import { useAchievements } from '@/org/use-achievements'
import { User } from '@/org/use-users'
import { Project } from '@/org/use-projects'

// Components
import GetStartedMenu from '@/components/GetStartedMenu.vue'

export default defineComponent({
  name: 'SystemBar',
  components: { GetStartedMenu },

  props: {
    user: {
      type: Object as PropType<User>,
      required: true,
    },
    project: {
      type: Object as PropType<Project>,
      required: true,
    },
  },

  setup(props) {
    const achievements = useAchievements(computed(() => props.project))
    return {
      achievements,
    }
  },
})
</script>

<style lang="scss" scoped>
a {
  color: map-get($grey, 'lighten-2') !important;
  font-weight: 600;
  text-decoration: none;

  &:hover {
    color: #fff !important;
    text-decoration: underline;
  }
}
</style>
