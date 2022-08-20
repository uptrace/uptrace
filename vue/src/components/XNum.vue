<script lang="ts">
import Vue from 'vue'

// Utilities
import { unitFromName, createFormatter, Unit } from '@/util/fmt'

export default Vue.component('XNum', {
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
      default: undefined,
    },
  },
  render(h, { props }) {
    const unit = props.unit ?? unitFromName(props.name, props.value)
    const fmt = createFormatter(unit)
    return h('span', fmt(props.value))
  },
})
</script>

<style lang="scss" scoped></style>
