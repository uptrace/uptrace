<template>
  <v-sheet ref="el" :dark="dark" class="x-code" :class="{ 'x-code--wrap': wrap }">
    <prism :code="code" :inline="inline" :language="language" />
  </v-sheet>
</template>

<script lang="ts">
import { defineComponent, ref, computed } from 'vue'

import Prism from 'vue-prism-component'

export default defineComponent({
  name: 'PrismCode',
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
    dark: {
      type: Boolean,
      default: false,
    },
  },

  setup(props) {
    const el = ref<any>()

    const wrap = computed((): boolean => {
      return !props.code.includes('\n')
    })

    return {
      el,
      wrap,
    }
  },
})
</script>

<style lang="scss">
.v-sheet.x-code {
  position: relative;
  margin: 16px 0;
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

  &.x-code--wrap {
    & code[class*='language'],
    & pre[class*='language'] {
      white-space: pre-wrap;
    }
  }

  &.theme--dark {
    code[class*='language'],
    pre[class*='language'] {
      color: #ccc !important;
    }

    pre[class*='language'] {
      &::after {
        color: hsla(0, 0%, 50%, 1);
      }
    }

    &.v-sheet--outlined {
      border: thin solid hsla(0, 0%, 100%, 0.12) !important;
    }

    .token.operator,
    .token.string {
      background: none;
    }

    .token.comment,
    .token.block-comment,
    .token.prolog,
    .token.doctype,
    .token.cdata {
      color: #999;
    }

    .token.punctuation {
      color: #ccc;
    }

    .token.tag,
    .token.attr-name,
    .token.namespace,
    .token.deleted {
      color: #e2777a;
    }

    .token.function-name {
      color: #6196cc;
    }

    .token.boolean,
    .token.number,
    .token.function {
      color: #f08d49;
    }

    .token.property,
    .token.class-name,
    .token.constant,
    .token.symbol {
      color: #f8c555;
    }

    .token.selector,
    .token.important,
    .token.atrule,
    .token.keyword,
    .token.builtin {
      color: #cc99cd;
    }

    .token.string,
    .token.char,
    .token.attr-value,
    .token.regex,
    .token.variable {
      color: #7ec699;
    }

    .token.operator,
    .token.entity,
    .token.url {
      color: #67cdcc;
    }

    .token.important,
    .token.bold {
      font-weight: bold;
    }

    .token.italic {
      font-style: italic;
    }

    .token.entity {
      cursor: help;
    }

    .token.inserted {
      color: green;
    }
  }
}
</style>
