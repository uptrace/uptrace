<template>
  <span v-frag>
    <template v-if="json">
      <slot name="code" :code="json">
        <PrismCode :code="json" class="my-0" />
      </slot>
    </template>
    <template v-else-if="isFormatted">
      <slot name="code" :code="value">
        <PrismCode :code="value" class="my-0" />
      </slot>
    </template>
    <template v-else>
      <slot name="text">
        <AnyValue :value="value" :name="name" />
      </slot>
    </template>
  </span>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

// Components
import PrismCode from '@/components/PrismCode.vue'
import AnyValue from '@/components/AnyValue.vue'

export default defineComponent({
  name: 'CodeOrText',
  components: { PrismCode, AnyValue },

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

  setup(props) {
    const json = computed(() => {
      if (!props.value) {
        return undefined
      }
      if (typeof props.value === 'object' && !Array.isArray(props.value)) {
        return prettyPrint(props.value)
      }
      if (typeof props.value !== 'string' || !isJson(props.value)) {
        return undefined
      }

      const obj = parseJSON(props.value)
      if (!obj) {
        return undefined
      }
      return prettyPrint(obj)
    })

    const isFormatted = computed(() => {
      return typeof props.value === 'string' && props.value.trim().includes('\n')
    })

    return { json, isFormatted }
  },
})

function prettyPrint(v: any): string {
  return JSON.stringify(v, null, 2)
}

function parseJSON(s: string): any {
  try {
    return JSON.parse(s)
  } catch (_) {
    return
  }
}

export function isJson(value: string): boolean {
  if (value.length < 2) {
    return false
  }

  const s = value.trim()
  const res = s[0] + s[s.length - 1]
  return res === '{}'
}
</script>

<style lang="scss" scoped></style>
