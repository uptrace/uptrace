<template>
  <v-autocomplete
    v-autowidth="{ minWidth: '40px' }"
    :value="value"
    :items="items"
    placeholder="none"
    :prefix="`${label}: `"
    multiple
    clearable
    auto-select-first
    hide-details
    dense
    outlined
    background-color="light-blue lighten-5"
    class="mr-2 fit"
    @change="$emit('input', $event)"
  >
  </v-autocomplete>
</template>

<script lang="ts">
import { defineComponent, shallowRef, computed, PropType } from 'vue'

// Composables
import { useRouteQuery } from '@/use/router'

export default defineComponent({
  name: 'StickyFilter',

  props: {
    value: {
      type: Array as PropType<string[]>,
      required: true,
    },
    loading: {
      type: Boolean,
      default: false,
    },
    items: {
      type: Array as PropType<string[]>,
      required: true,
    },
    label: {
      type: String,
      required: true,
    },
  },

  setup(props, ctx) {
    const menu = shallowRef(false)

    const attrs = computed(() => {
      if (props.outlined) {
        return { outlined: true }
      }
      return { dark: true, class: 'blue darken-1 elevation-5' }
    })

    useRouteQuery().sync({
      fromQuery(params) {
        if (!Object.keys(params).length) {
          return
        }

        if (!params.env) {
          ctx.emit('input', [])
          return
        }

        if (Array.isArray(params.env)) {
          ctx.emit('input', params.env)
        } else if (typeof params.env === 'string') {
          ctx.emit('input', [params.env])
        }
      },
      toQuery() {
        if (props.value.length) {
          return { env: props.value }
        }
        return {}
      },
    })

    return {
      menu,
      attrs,
    }
  },
})
</script>

<style lang="scss" scoped>
.fit {
  flex: fit-content 0 0 !important;
}
</style>
