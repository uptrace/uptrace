<script lang="ts">
import { VNode } from 'vue'
import { defineComponent, h } from '@vue/composition-api'

export default defineComponent({
  name: 'XPlaceholder',

  setup(props, ctx) {
    return () => {
      if (ctx.slots.placeholder) {
        return singleNode(ctx.slots.placeholder())
      }
      if (ctx.slots.default) {
        return singleNode(ctx.slots.default())
      }
      return
    }
  },
})

function singleNode(nodes: VNode[] | VNode) {
  if (Array.isArray(nodes) && nodes.length > 1) {
    return h('div', nodes)
  }
  return nodes
}
</script>
