<script lang="ts">
import Vue from 'vue'

// Misc
import { duration } from '@/util/fmt'

export default Vue.component('DurationValue', {
  functional: true,
  props: {
    value: {
      type: Number,
      default: undefined,
    },
  },

  render(h, { props }) {
    const text = duration(props.value)
    return h('span', text)
  },
})
</script>

<style lang="scss" scoped></style>
