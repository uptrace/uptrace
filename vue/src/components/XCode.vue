<template>
  <v-sheet class="x-code" :class="{ 'x-code--wrap': wrap }">
    <prism :code="code" :inline="inline" :language="language" />
  </v-sheet>
</template>

<script lang="ts">
import { defineComponent, computed } from 'vue'

import Prism from 'vue-prism-component'

export default defineComponent({
  name: 'XCode',
  components: { Prism },

  props: {
    code: {
      type: String,
      required: true,
    },
    language: {
      type: String,
      default: 'markup',
    },
    inline: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const wrap = computed((): boolean => {
      return !props.code.includes('\n')
    })

    return { wrap }
  },
})
</script>

<style lang="scss">
.v-sheet.x-code {
  margin: 16px 0;
  position: relative;
  padding: 12px;
  background-color: map-get($grey, 'lighten-5');

  pre,
  code {
    background: transparent;
    font-size: 1rem;
    font-weight: 300;
    margin: 0 !important;
  }

  > pre {
    border-radius: inherit;
  }

  &.x-code--wrap {
    & code[class*='language'],
    & pre[class*='language'] {
      white-space: pre-wrap;
    }
  }

  code[class*='language'],
  pre[class*='language'] {
    background: none;
    font-family: Consolas, Monaco, 'Andale Mono', 'Ubuntu Mono', monospace;
    font-size: 0.875rem;
    hyphens: none;
    line-height: 1.5;
    margin: 0;
    padding: 0;
    tab-size: 4;
    text-align: left;
    text-shadow: none;
    white-space: pre;
    word-break: normal;
    word-spacing: normal;
    word-wrap: normal;
  }
}
</style>
