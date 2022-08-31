<script lang="ts">
import Vue from 'vue'

const BREAK_MIN_LEN = 0

export default Vue.component('AnyValue', {
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

    if (typeof props.value !== 'string') {
      return h('span', data, String(props.value))
    }

    if (props.value.length >= BREAK_MIN_LEN) {
      data = { ...data, class: [data.class, 'word-break-all'] }
    }
    return h('span', data, props.value)
  },
})
</script>

<style lang="scss" scoped></style>
