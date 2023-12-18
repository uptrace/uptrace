<template>
  <v-menu offset-y>
    <template #activator="{ on, attrs }">
      <v-btn icon v-bind="attrs" v-on="on">
        <v-icon>mdi-dots-vertical</v-icon>
      </v-btn>
    </template>
    <v-list>
      <WhereMenu :uql="uql" :axios-params="axiosParams" />
      <AggByMenu :uql="uql" :axios-params="axiosParams" :disabled="aggDisabled" />
      <GroupByMenu :uql="uql" :axios-params="axiosParams" :disabled="aggDisabled" />
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { UseUql } from '@/use/uql'

// Components
import WhereMenu from '@/tracing/query/WhereMenu.vue'
import AggByMenu from '@/tracing/query/AggByMenu.vue'
import GroupByMenu from '@/tracing/query/GroupByMenu.vue'

export default defineComponent({
  name: 'AdvancedMenu',
  components: { WhereMenu, AggByMenu, GroupByMenu },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    axiosParams: {
      type: Object as PropType<Record<string, any>>,
      required: true,
    },
    aggDisabled: {
      type: Boolean,
      default: false,
    },
  },
})
</script>

<style lang="scss" scoped></style>
