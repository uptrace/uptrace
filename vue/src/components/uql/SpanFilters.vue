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
    <AggFilterMenu
      :uql="uql"
      :axios-params="axiosParams"
      :disabled="$route.name !== groupListRoute"
    />
    <GroupByMenu
      :uql="uql"
      :axios-params="axiosParams"
      :disabled="$route.name !== groupListRoute"
    />

    <v-divider vertical class="mx-2" />
    <v-btn text class="v-btn--filter" @click="$emit('click:reset')">Reset</v-btn>
    <v-btn text class="v-btn--filter" @click="uql.rawMode = true">Edit</v-btn>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { AxiosParams } from '@/use/axios'
import { UseSystems } from '@/use/systems'
import { UseUql } from '@/use/uql'

// Components
import SearchFilterMenu from '@/components/uql/SearchFilterMenu.vue'
import DurationFilterMenu from '@/components/uql/DurationFilterMenu.vue'
import AttrFilterMenu from '@/components/uql/AttrFilterMenu.vue'
import WhereFilterMenu from '@/components/uql/WhereFilterMenu.vue'
import AggFilterMenu from '@/components/uql/AggFilterMenu.vue'
import GroupByMenu from '@/components/uql/GroupByMenu.vue'

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
    groupListRoute: {
      type: String,
      required: true,
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
