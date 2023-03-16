<template>
  <div v-frag>
    <SearchFilterMenu :uql="uql" />
    <DurationFilterMenu :uql="uql" />
    <AttrFilterMenu
      :uql="uql"
      :axios-params="axiosParams"
      :attr-key="AttrKey.spanStatusCode"
      label="Status"
    />
    <v-btn text class="v-btn--filter" @click="drawer = true">
      Filters
      <v-icon color="green">mdi-new-box</v-icon>
    </v-btn>

    <v-divider vertical class="mx-2" />

    <WhereFilterMenu :systems="systems" :uql="uql" :axios-params="axiosParams" />
    <AggByMenu :uql="uql" :axios-params="axiosParams" :disabled="aggDisabled" />
    <GroupByMenu :uql="uql" :axios-params="axiosParams" :disabled="aggDisabled" />

    <v-divider vertical class="mx-2" />
    <SpanQueryHelpDialog />
    <v-btn text class="v-btn--filter" @click="$emit('click:reset')">Reset</v-btn>
    <v-btn text class="v-btn--filter" @click="uql.rawMode = true">Edit</v-btn>

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
        component="tracing"
        :uql="uql"
        :axios-params="facetParams"
        @input="drawer = $event"
      />
    </v-navigation-drawer>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { AxiosParams } from '@/use/axios'
import { UseSystems } from '@/tracing/system/use-systems'
import { UseUql } from '@/use/uql'

// Components
import SearchFilterMenu from '@/tracing/query/SearchFilterMenu.vue'
import DurationFilterMenu from '@/tracing/query/DurationFilterMenu.vue'
import AttrFilterMenu from '@/tracing/query/AttrFilterMenu.vue'
import WhereFilterMenu from '@/tracing/query/WhereFilterMenu.vue'
import AggByMenu from '@/tracing/query/AggByMenu.vue'
import GroupByMenu from '@/tracing/query/GroupByMenu.vue'
import SpanQueryHelpDialog from '@/tracing/query/SpanQueryHelpDialog.vue'
import FacetList from '@/components/facet/FacetList.vue'

// Utilities
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'SpanQueryBuilder',
  components: {
    SearchFilterMenu,
    DurationFilterMenu,
    AttrFilterMenu,
    WhereFilterMenu,
    AggByMenu,
    GroupByMenu,
    SpanQueryHelpDialog,
    FacetList,
  },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<AxiosParams>,
      required: true,
    },
    aggDisabled: {
      type: Boolean,
      default: false,
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

    return { AttrKey, drawer, facetParams, onClickOutside, closeConditional }
  },
})
</script>

<style lang="scss" scoped>
.v-divider--vertical {
  margin-top: 6px;
  margin-bottom: 4px;
}
</style>
