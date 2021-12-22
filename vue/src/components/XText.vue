<script lang="ts">
import Vue from 'vue'

import XNum from '@/components/XNum.vue'
import XCode from '@/components/XCode.vue'

export default Vue.component('XText', {
  functional: true,
  props: {
    name: {
      type: String,
      default: '',
    },
    value: {
      type: undefined,
      required: true,
    },
  },
  render(h, { props, data }) {
    if (Array.isArray(props.value)) {
      return h('span', data, props.value.join(', '))
    }

    switch (typeof props.value) {
      case 'object':
        return h(XCode, {
          props: {
            code: JSON.stringify(props.value, null, 2),
          },
        })
      case 'string': {
        const obj = parseJSON(props.value)
        if (typeof obj === 'object') {
          return h(XCode, {
            props: {
              code: JSON.stringify(obj, null, 2),
            },
          })
        }
        break
      }
    }

    if (isFormatted(props.value)) {
      return h(XCode, {
        props: {
          code: props.value,
        },
      })
    }

    return h(XNum, {
      props: {
        name: props.name,
        value: props.value,
      },
    })
  },
})

function parseJSON(s: string): any {
  try {
    return JSON.parse(s)
  } catch (_) {
    return
  }
}

export function isFormatted(value: unknown): value is string {
  if (typeof value !== 'string') {
    return false
  }
  return value.includes('\n')
}
</script>

<style lang="scss" scoped></style>
