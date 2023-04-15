<script lang="ts">
import Vue from 'vue'
import { PropType } from 'vue'

// Utilities
import { percents } from '@/util/fmt/num'
import { createFormatter, Unit } from '@/util/fmt'
import { formatTemplate } from '@/util/string'

export default Vue.component('XPct', {
  functional: true,
  props: {
    a: {
      type: Number,
      required: true,
    },
    b: {
      type: Number,
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
    title: {
      type: String,
      default: '{0} of {1}',
    },
  },
  render(h, { props, data }) {
    const fmt = createFormatter(props.unit)
    data.attrs = {
      ...data.attrs,
      title: formatTemplate(props.title, fmt(props.a), fmt(props.b)),
    }
    return h('span', data, percents(pct(props.a, props.b)))
  },
})

function pct(a: number, b: number): number {
  if (a === 0 || b === 0) {
    return 0
  }

  const pct = a / b
  if (pct > 1 || pct === Infinity) {
    return 1
  }
  if (pct === -Infinity) {
    return -1
  }
  return pct
}
</script>

<style lang="scss" scoped></style>
