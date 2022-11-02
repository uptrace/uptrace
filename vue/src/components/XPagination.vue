<template>
  <v-row v-if="pager.numPage > 1" align="center" justify="space-between" no-gutters>
    <v-col v-if="showPager" cols="3"></v-col>
    <v-col>
      <v-pagination
        v-model="pager.page"
        :total-visible="totalVisible"
        :length="pager.numPage"
      ></v-pagination>
    </v-col>
    <v-col v-if="showPager" cols="3" class="text-right">
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
    totalVisible: {
      type: [String, Number],
      default: 7,
    },
    showPager: {
      type: Boolean,
      default: true,
    },
  },
})
</script>

<style lang="scss" scoped></style>
