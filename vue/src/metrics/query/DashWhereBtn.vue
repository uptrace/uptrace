<template>
  <span>
    <v-btn text class="v-btn--filter" @click="drawer = !drawer">
      <v-icon left>mdi-filter</v-icon>
      <span>Where</span>
    </v-btn>

    <v-navigation-drawer
      v-model="drawer"
      v-click-outside="{
        handler: onClickOutside,
        closeConditional,
      }"
      app
      right
      :width="width"
      :temporary="temporary"
      stateless
    >
      <v-system-bar window>
        <v-btn
          icon
          :title="temporary ? 'Keep menu open' : 'Hide menu'"
          @click="temporary = !temporary"
        >
          <v-icon>{{ temporary ? 'mdi-dock-right' : 'mdi-dock-window' }}</v-icon>
        </v-btn>
        <v-btn v-if="!temporary" icon title="Close menu" @click="drawer = false">
          <v-icon>mdi-close</v-icon>
        </v-btn>

        <v-btn-toggle v-model="width" group dense class="ml-4">
          <v-btn :value="300" icon>
            <v-icon>mdi-size-s</v-icon>
          </v-btn>
          <v-btn :value="400" icon>
            <v-icon>mdi-size-m</v-icon>
          </v-btn>
          <v-btn :value="500" icon>
            <v-icon>mdi-size-l</v-icon>
          </v-btn>
        </v-btn-toggle>
      </v-system-bar>

      <FacetList
        component="metrics"
        :uql="uql"
        :axios-params="facetParams"
        @input="drawer = $event"
      />
    </v-navigation-drawer>
  </span>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { useStorage } from '@/use/local-storage'
import { UseUql } from '@/use/uql'

// Components
import FacetList from '@/components/facet/FacetList.vue'

export default defineComponent({
  name: 'DashWhereBtn',
  components: { FacetList },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
  },

  setup(props) {
    const drawer = shallowRef(false)
    const width = useStorage('tracing-width', 400)
    const temporary = useStorage('tracing-temporary', true)

    const facetParams = computed(() => {
      if (!drawer.value) {
        return null
      }
      return {
        ...props.axiosParams,
        query: props.uql.whereQuery,
      }
    })

    function onClickOutside() {
      if (temporary.value) {
        drawer.value = false
      }
    }

    function closeConditional() {
      return drawer.value
    }

    return {
      facetParams,

      drawer,
      width,
      temporary,

      onClickOutside,
      closeConditional,
    }
  },
})
</script>

<style lang="scss" scoped></style>
