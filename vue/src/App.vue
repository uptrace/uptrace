<template>
  <v-app>
    <v-app-bar app absolute flat color="white" class="v-bar--underline">
      <v-container fluid class="pa-0 fill-height">
        <v-row align="center" class="flex-nowrap">
          <v-col cols="auto">
            <div class="mt-2">
              <UptraceLogoLarge
                v-if="!user.isAuth || $vuetify.breakpoint.lgAndUp"
                :to="{ name: 'Home' }"
              />
              <UptraceLogoSmall v-else :to="{ name: 'Home' }" />
            </div>
          </v-col>

          <v-col cols="auto">
            <ProjectPicker />
          </v-col>

          <v-spacer v-if="$vuetify.breakpoint.mdAndDown" />

          <v-col cols="auto">
            <v-tabs optional class="ml-lg-10 ml-xl-16">
              <template v-if="user.isAuth && $route.params.projectId">
                <v-tab :to="{ name: 'Overview' }">Overview</v-tab>
                <v-tab :to="{ name: 'SpanGroupList' }">Tracing</v-tab>
                <v-tab :to="{ name: 'MetricsDashList' }">Metrics</v-tab>
              </template>
              <v-tab v-if="!user.isAuth" :to="{ name: 'Login' }">Login</v-tab>
            </v-tabs>
          </v-col>

          <v-spacer />

          <v-col v-if="user.isAuth && $route.params.projectId" cols="auto">
            <Search />
          </v-col>
          <v-col cols="auto">
            <v-btn v-if="user.isAuth && $route.params.projectId" :to="helpRoute" icon>
              <v-icon>mdi-help-circle</v-icon>
            </v-btn>

            <v-menu v-if="user.isAuth" bottom offset-y>
              <template #activator="{ on }">
                <v-btn icon v-on="on">
                  <v-icon>mdi-dots-vertical</v-icon>
                </v-btn>
              </template>

              <v-list>
                <v-list-item>
                  <v-list-item-content>
                    <v-list-item-title class="font-weight-bold">{{
                      user.current.username || 'Anonymous'
                    }}</v-list-item-title>
                  </v-list-item-content>
                </v-list-item>

                <v-divider></v-divider>

                <v-list-item @click="user.logout">
                  <v-list-item-title>Sign out</v-list-item-title>
                </v-list-item>
              </v-list>
            </v-menu>
          </v-col>
        </v-row>
      </v-container>
    </v-app-bar>

    <v-main>
      <XSnackbar />
      <router-view :date-range="dateRange" />
    </v-main>

    <v-footer app absolute color="grey lighten-5">
      <v-container fluid>
        <v-row justify="center" align="center">
          <v-col cols="auto">
            <v-btn href="https://uptrace.dev/get/" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-help-circle-outline</v-icon>
              <span>Docs</span>
            </v-btn>
            <v-btn href="https://uptrace.dev/opentelemetry/" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-open-source-initiative</v-icon>
              <span>OpenTelemetry</span>
            </v-btn>
            <v-btn
              href="https://uptrace.dev/opentelemetry/instrumentations/"
              target="_blank"
              text
              rounded
              small
            >
              <v-icon small class="mr-1">mdi-toy-brick-outline</v-icon>
              <span>Instrumentations</span>
            </v-btn>
            <v-btn href="https://t.me/uptrace" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-message-outline</v-icon>
              <span>Chat</span>
            </v-btn>
            <v-btn href="https://github.com/uptrace/uptrace" target="_blank" text rounded small>
              <v-icon small class="mr-1">mdi-github</v-icon>
              <span>GitHub</span>
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
    </v-footer>
  </v-app>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Composables
import { useRoute, useRouteQuery } from '@/use/router'
import { useForceReload } from '@/use/force-reload'
import { useDateRange } from '@/use/date-range'
import { useUser } from '@/use/org'

// Components
import UptraceLogoLarge from '@/components/UptraceLogoLarge.vue'
import UptraceLogoSmall from '@/components/UptraceLogoSmall.vue'
import ProjectPicker from '@/components/ProjectPicker.vue'
import Search from '@/components/Search.vue'
import XSnackbar from '@/components/XSnackbar.vue'

// Utilities
import { SystemName } from '@/models/otel'

export default defineComponent({
  name: 'App',
  components: { UptraceLogoLarge, UptraceLogoSmall, ProjectPicker, Search, XSnackbar },

  setup() {
    useRouteQuery()
    useForceReload()
    const dateRange = useDateRange()

    const route = useRoute()
    const user = useUser()

    const helpRoute = computed(() => {
      if (route.value.name && route.value.name.startsWith('Metrics')) {
        return {
          name: 'MetricsHelp',
        }
      }
      return {
        name: 'TracingHelp',
        params: { projectId: route.value.params.projectId },
      }
    })

    return {
      SystemName,
      dateRange,
      user,
      helpRoute,
    }
  },
})
</script>

<style lang="scss">
.theme--light,
.theme--dark {
  .v-bar--underline {
    border-width: 0 0 thin 0;
    border-style: solid;

    &.theme--light {
      border-bottom-color: #0000001f !important;
    }

    &.theme--dark {
      border-bottom-color: #ffffff1f !important;
    }
  }
}
</style>

<style lang="scss" scoped>
.v-footer strong {
  font-weight: 500;
}

.v-app-bar ::v-deep .v-toolbar__content {
  padding-top: 0px;
  padding-bottom: 0px;
}
</style>
