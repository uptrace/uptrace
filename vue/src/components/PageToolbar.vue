<template>
  <div class="page-toolbar bg--light-primary">
    <v-progress-linear v-if="loading" absolute indeterminate></v-progress-linear>
    <v-container :fluid="localFluid" class="py-0">
      <v-toolbar color="transparent" flat height="auto">
        <slot></slot>

        <template v-if="$slots.extension" #extension>
          <slot name="extension"></slot>
        </template>
      </v-toolbar>
    </v-container>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'
import vuetify from '@/plugins/vuetify'

export default defineComponent({
  name: 'PageToolbar',

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    fluid: {
      type: Boolean,
      default: undefined,
    },
  },

  setup(props) {
    const fluid = computed(() => {
      if (props.fluid !== undefined) {
        return props.fluid
      }
      return vuetify.framework.breakpoint.mdAndDown
    })

    return { localFluid: fluid }
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
