<template>
  <v-btn
    absolute
    class="v-btn--copy"
    icon
    right
    style="background-color: inherit"
    top
    @click="copy"
  >
    <v-fade-transition hide-on-leave>
      <v-icon :key="String(clicked)" color="grey">{{
        clicked ? 'mdi-check' : 'mdi-content-copy'
      }}</v-icon>
    </v-fade-transition>
  </v-btn>
</template>

<script lang="ts">
import { defineComponent, shallowRef, PropType } from 'vue'

export default defineComponent({
  name: 'CopyBtn',

  props: {
    target: {
      type: Function as PropType<() => HTMLElement>,
      required: true,
    },
  },

  setup(props) {
    const clicked = shallowRef(false)

    async function copy() {
      const el = props.target()

      el.setAttribute('contenteditable', 'true')
      el.focus()

      document.execCommand('selectAll', false)
      document.execCommand('copy')
      removeSelection()

      el.removeAttribute('contenteditable')

      clicked.value = true

      await wait(2000)

      clicked.value = false
    }

    return { clicked, copy }
  },
})

function removeSelection() {
  if (!window.getSelection) {
    return
  }

  const sel = window.getSelection()
  if (!sel) {
    return
  }

  if (sel.empty) {
    // Chrome
    sel.empty()
    return
  }

  if (sel.removeAllRanges) {
    // Firefox
    sel.removeAllRanges()
    return
  }
}

function wait(timeout: number) {
  return new Promise((resolve) => setTimeout(resolve, timeout))
}
</script>
