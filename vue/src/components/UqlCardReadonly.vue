<template>
  <div>
    <UptraceQueryChip v-for="(part, i) in parts" :key="i" :query="part" class="mr-2 mb-1" />
  </div>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Composables
import { splitQuery } from '@/use/uql'

// Components
import UptraceQueryChip from '@/components/UptraceQueryChip.vue'

export default defineComponent({
  name: 'UqlCardReadonly',
  components: { UptraceQueryChip },

  props: {
    query: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const parts = computed(() => {
      return splitQuery(props.query)
    })

    return { parts }
  },
})
</script>

<style lang="scss" scoped></style>
