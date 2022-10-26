<template>
  <div v-frag>
    <SearchFilterMenu :uql="uql" />
    <DurationFilterMenu :uql="uql" />
    <AttrFilterMenu
      :uql="uql"
      :axios-params="axiosParams"
      :attr-key="xkey.spanStatusCode"
      label="Status"
    />
    <AttrFilterMenu :uql="uql" :axios-params="axiosParams" :attr-key="xkey.spanKind" label="Kind" />

    <v-divider vertical class="mx-2" />

    <WhereFilterMenu :systems="systems" :uql="uql" :axios-params="axiosParams" />
    <AggFilterMenu :uql="uql" :axios-params="axiosParams" :disabled="aggDisabled" />
    <GroupByMenu :uql="uql" :axios-params="axiosParams" :disabled="aggDisabled" />

    <v-divider vertical class="mx-2" />
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
import SearchFilterMenu from '@/tracing/uql/SearchFilterMenu.vue'
import DurationFilterMenu from '@/tracing/uql/DurationFilterMenu.vue'
import AttrFilterMenu from '@/tracing/uql/AttrFilterMenu.vue'
import WhereFilterMenu from '@/tracing/uql/WhereFilterMenu.vue'
import AggFilterMenu from '@/tracing/uql/AggFilterMenu.vue'
import GroupByMenu from '@/tracing/uql/GroupByMenu.vue'

// Utilities
import { xkey } from '@/models/otelattr'

export default defineComponent({
  name: 'SpanFilters',
  components: {
    SearchFilterMenu,
    DurationFilterMenu,
    AttrFilterMenu,
    WhereFilterMenu,
    AggFilterMenu,
    GroupByMenu,
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
    return { xkey }
  },
})
</script>

<style lang="scss" scoped>
.v-divider--vertical {
  margin-top: 6px;
  margin-bottom: 4px;
}
</style>
