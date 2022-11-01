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
    <AttrFilterMenu
      :uql="uql"
      :axios-params="axiosParams"
      :attr-key="AttrKey.spanKind"
      label="Kind"
    />

    <v-divider vertical class="mx-2" />

    <WhereFilterMenu :systems="systems" :uql="uql" :axios-params="axiosParams" />
    <AggFilterMenu :uql="uql" :axios-params="axiosParams" :disabled="aggDisabled" />
    <GroupByMenu :uql="uql" :axios-params="axiosParams" :disabled="aggDisabled" />

    <v-divider vertical class="mx-2" />
    <SpanQueryHelpDialog />
    <v-btn text class="v-btn--filter" @click="$emit('click:reset')">Reset</v-btn>
    <v-btn text class="v-btn--filter" @click="uql.rawMode = true">Edit</v-btn>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { AxiosParams } from '@/use/axios'
import { UseSystems } from '@/tracing/use-systems'
import { UseUql } from '@/use/uql'

// Components
import SearchFilterMenu from '@/tracing/query/SearchFilterMenu.vue'
import DurationFilterMenu from '@/tracing/query/DurationFilterMenu.vue'
import AttrFilterMenu from '@/tracing/query/AttrFilterMenu.vue'
import WhereFilterMenu from '@/tracing/query/WhereFilterMenu.vue'
import AggFilterMenu from '@/tracing/query/AggFilterMenu.vue'
import GroupByMenu from '@/tracing/query/GroupByMenu.vue'
import SpanQueryHelpDialog from '@/tracing/query/SpanQueryHelpDialog.vue'

// Utilities
import { AttrKey } from '@/models/otelattr'

export default defineComponent({
  name: 'SpanQueryBuilder',
  components: {
    SearchFilterMenu,
    DurationFilterMenu,
    AttrFilterMenu,
    WhereFilterMenu,
    AggFilterMenu,
    GroupByMenu,
    SpanQueryHelpDialog,
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

  setup() {
    return { AttrKey }
  },
})
</script>

<style lang="scss" scoped>
.v-divider--vertical {
  margin-top: 6px;
  margin-bottom: 4px;
}
</style>
