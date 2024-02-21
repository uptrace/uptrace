<template>
  <div class="d-flex">
    <WhereSidebarBtn :uql="uql" :axios-params="axiosParams" />
    <SearchFilterMenu :systems="systems" :uql="uql" />
    <DurationFilterMenu v-if="systems.isSpan" :uql="uql" />
    <AttrFilterMenu
      :uql="uql"
      :axios-params="axiosParams"
      :attr-key="AttrKey.spanStatusCode"
      label="Status"
    />
    <AdvancedMenu :uql="uql" :axios-params="axiosParams" :agg-disabled="aggDisabled" />

    <v-divider vertical class="mx-2" />

    <QueryHelpDialog />
    <v-btn text class="v-btn--filter" @click="$emit('click:reset')">Reset</v-btn>
    <v-btn text class="v-btn--filter" @click="uql.rawMode = true">Edit</v-btn>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { UseSystems } from '@/tracing/system/use-systems'
import { UseUql } from '@/use/uql'

// Components
import SearchFilterMenu from '@/tracing/query/SearchFilterMenu.vue'
import DurationFilterMenu from '@/tracing/query/DurationFilterMenu.vue'
import AttrFilterMenu from '@/tracing/query/AttrFilterMenu.vue'
import AdvancedMenu from '@/tracing/query/AdvancedMenu.vue'
import QueryHelpDialog from '@/tracing/query/QueryHelpDialog.vue'
import WhereSidebarBtn from '@/tracing/query/WhereSidebarBtn.vue'

// Misc
import { AttrKey } from '@/models/otel'

export default defineComponent({
  name: 'SpanQueryBuilder',
  components: {
    SearchFilterMenu,
    DurationFilterMenu,
    AttrFilterMenu,
    AdvancedMenu,
    QueryHelpDialog,
    WhereSidebarBtn,
  },

  props: {
    systems: {
      type: Object as PropType<UseSystems>,
      required: true,
    },
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    axiosParams: {
      type: Object,
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

<style lang="scss" scoped></style>
