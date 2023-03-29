<template>
  <span>
    <v-btn text class="v-btn--filter" @click="drawer = true">
      <v-icon left>mdi-page-layout-sidebar-left</v-icon>
      <span>Where</span>
    </v-btn>

    <v-navigation-drawer
      v-model="drawer"
      v-click-outside="{
        handler: onClickOutside,
        closeConditional,
      }"
      app
      temporary
      stateless
      width="500"
    >
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
      drawer.value = false
    }

    function closeConditional() {
      return drawer.value
    }

    return {
      facetParams,

      drawer,
      onClickOutside,
      closeConditional,
    }
  },
})
</script>

<style lang="scss" scoped></style>
