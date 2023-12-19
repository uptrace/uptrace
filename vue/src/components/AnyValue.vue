<script lang="ts">
import Vue from 'vue'
import { parse } from 'date-fns'

// Components
import DateValue from '@/components/DateValue.vue'

// Misc
import { Unit } from '@/util/fmt'

const BREAK_MIN_LEN = 0

export default Vue.component('AnyValue', {
  functional: true,
  props: {
    value: {
      type: undefined,
      required: true,
    },
    unit: {
      type: String,
      default: '',
    },
  },
  render(h, { props, data }) {
    if (props.value === null) {
      return h('span', data, '<null>')
    }
    if (props.value === '') {
      return h('span', data, '<empty>')
    }

    if (Array.isArray(props.value)) {
      const str = props.value.join(', ')
      if (str.length >= BREAK_MIN_LEN) {
        data = { ...data, class: [data.class, 'word-break-all'] }
      }
      return h('span', data, str)
    }

    if (typeof props.value === 'number') {
      switch (props.unit) {
        case Unit.UnixTime:
          const unix = props.value as number
          return h(DateValue, {
            props: {
              value: new Date(unix * 1000),
            },
          })
      }
    }

    if (typeof props.value !== 'string') {
      return h('span', data, String(props.value))
    }

    const date = parse(props.value, "yyyy-MM-dd'T'HH:mm:ssXXX", new Date())
    if (!isNaN(date as any)) {
      return h(DateValue, {
        props: {
          value: date,
        },
      })
    }

    if (props.value.length >= BREAK_MIN_LEN) {
      data = { ...data, class: [data.class, 'word-break-all'] }
    }
    return h('span', data, props.value)
  },
})
</script>

<style lang="scss" scoped></style>
