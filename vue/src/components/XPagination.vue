<template>
  <v-row v-if="pager.numPage > 1" align="center" justify="space-between" no-gutters>
    <v-col cols="2"></v-col>
    <v-col>
      <v-pagination v-model="pager.page" :total-visible="7" :length="pager.numPage"></v-pagination>
    </v-col>
    <v-col cols="2" class="text-right">
      <XPager v-if="pager.numItem" :pager="pager" without-prev-next />
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'

// Composables
import { UsePager } from '@/use/pager'

// Components
import XPager from '@/components/XPager.vue'

export default defineComponent({
  name: 'XPagination',
  components: { XPager },

  props: {
    pager: {
      type: Object as PropType<UsePager>,
      required: true,
    },
  },
})
</script>

<style lang="scss" scoped></style>
