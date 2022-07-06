<template>
  <div class="page-toolbar" :class="color">
    <v-progress-linear v-if="loading" absolute indeterminate></v-progress-linear>
    <v-container :fluid="fluid" class="py-0">
      <v-row class="ma-0">
        <v-col class="pa-0">
          <v-toolbar flat color="transparent" height="auto">
            <slot></slot>
          </v-toolbar>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'PageToolbar',

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    fluid: {
      type: Boolean,
      default: false,
    },
    color: {
      type: String,
      default: 'light-blue lighten-5',
    },
  },

  setup() {
    return {}
  },
})
</script>

<style lang="scss">
.page-toolbar {
  .v-breadcrumbs {
    padding: 0px;
  }

  .v-toolbar__content {
    min-height: 64px;
    padding-left: 0;
    padding-right: 0;

    .v-btn.v-btn--icon.v-size--default {
      height: 36px;
      width: 36px;
    }
  }

  .v-toolbar__title {
    padding-left: 8px;
    padding-right: 8px;
  }
}
</style>
