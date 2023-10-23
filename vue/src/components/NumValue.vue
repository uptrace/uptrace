<script lang="ts">
import Vue, { PropType } from 'vue'

// Utilities
import { createFormatter, Unit } from '@/util/fmt'

export default Vue.component('NumValue', {
  functional: true,
  props: {
    value: {
      type: undefined,
      required: true,
    },
    name: {
      type: String,
      default: '',
    },
    unit: {
      type: String as PropType<Unit>,
      default: Unit.None,
    },
  },
  render(h, { props }) {
    const fmt = createFormatter(props.unit)
    return h('span', fmt(props.value))
  },
})
</script>

<style lang="scss" scoped></style>
