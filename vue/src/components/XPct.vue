<script lang="ts">
import Vue from 'vue'
import { PropType } from 'vue'

// Utilities
import { percent } from '@/util/fmt/num'
import { unitFromName, createFormatter, Unit } from '@/util/fmt'
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
      default: undefined,
    },
    title: {
      type: String,
      default: '{0} of {1}',
    },
  },
  render(h, { props, data }) {
    const unit = props.unit ?? unitFromName(props.name, 0)

    const fmt = createFormatter(unit)

    data.attrs = {
      ...data.attrs,
      title: formatTemplate(props.title, fmt(props.a), fmt(props.b)),
    }

    return h('span', data, percent(pct(props.a, props.b)))
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
