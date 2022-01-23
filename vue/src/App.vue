<template>
  <v-app>
    <v-app-bar app elevation="1" color="white">
      <v-container :fluid="$vuetify.breakpoint.lgAndDown" class="pa-0 fill-height">
        <v-row align="center" class="flex-nowrap">
          <v-col cols="auto">
            <div class="mt-2">
              <UptraceLogoLarge v-if="$vuetify.breakpoint.lgAndUp" :to="{ name: 'Home' }" />
              <UptraceLogoSmall v-else :to="{ name: 'Home' }" />
            </div>
          </v-col>

          <v-col cols="auto">
            <ProjectPicker />
          </v-col>

          <v-col cols="auto">
            <v-tabs optional class="ml-8">
              <v-tab :to="{ name: 'Overview' }">Overview</v-tab>
              <v-tab :to="{ name: 'GroupList' }">Explore</v-tab>
              <v-tab :to="{ name: 'Help' }">Help</v-tab>
            </v-tabs>
          </v-col>

          <v-spacer></v-spacer>

          <v-col cols="auto">
            <v-text-field
              v-model="traceId"
              prepend-inner-icon="mdi-magnify"
              placeholder="Jump to trace id..."
              hide-details
              flat
              solo
              background-color="grey lighten-4"
              style="min-width: 300px; width: 300px"
              @keyup.enter="jumpToTrace"
            />
          </v-col>
          <v-col cols="auto" class="d-none d-md-inline-block">
            <v-btn
              href="https://uptrace.dev/compare-open-source"
              target="_blank"
              dark
              class="deep-orange darken-3 ml-lg-10"
              >Upgrade</v-btn
            >
          </v-col>
        </v-row>
      </v-container>
    </v-app-bar>

    <v-main>
      <XSnackbar />
      <router-view />
    </v-main>

    <v-footer app absolute color="grey lighten-5">
      <v-container fluid>
        <v-row justify="center" align="center">
          <v-col cols="auto">
            Found a bug? Create an
            <a href="https://github.com/uptrace/uptrace/issues" target="_blank">issue</a>. Have a
            question? Start a
            <a href="https://github.com/uptrace/uptrace/discussions" target="_blank">discussion</a>.
          </v-col>
        </v-row>
      </v-container>
    </v-footer>
  </v-app>
</template>

<script lang="ts">
import { defineComponent, shallowRef } from '@vue/composition-api'

// Composables
import { useRouter, useQuery } from '@/use/router'

// Components
import UptraceLogoLarge from '@/components/UptraceLogoLarge.vue'
import UptraceLogoSmall from '@/components/UptraceLogoSmall.vue'
import ProjectPicker from '@/components/ProjectPicker.vue'
import XSnackbar from '@/components/XSnackbar.vue'

export default defineComponent({
  name: 'App',
  components: { UptraceLogoLarge, UptraceLogoSmall, ProjectPicker, XSnackbar },

  setup() {
    const { router } = useRouter()
    useQuery()
    const traceId = shallowRef('')

    function jumpToTrace() {
      if (traceId.value) {
        router.push({ name: 'TraceShow', params: { traceId: traceId.value } })
        traceId.value = ''
      }
    }

    return { traceId, jumpToTrace }
  },
})
</script>

<style lang="scss" scoped>
.v-app-bar ::v-deep .v-toolbar__content {
  padding-top: 0px;
  padding-bottom: 0px;
}
</style>
