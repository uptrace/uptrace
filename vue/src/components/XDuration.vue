<script lang="ts">
import Vue from 'vue'

// Utilities
import { durationFixed } from '@/util/fmt'

export default Vue.component('XDuration', {
  functional: true,
  props: {
    duration: {
      type: Number,
      default: undefined,
    },
  },

  render(h, { props }) {
    const text = durationFixed(props.duration)
    return h('span', text)
  },
})
</script>

<style lang="scss" scoped></style>
