<script lang="ts">
import Vue from 'vue'
import { VIcon } from 'vuetify/lib'
import { PropType } from '@vue/composition-api'

// Composables
import { UseOrder } from '@/use/order'

// Styles
import 'vuetify/src/components/VDataTable/VDataTableHeader.sass'

enum Align {
  Start = 'start',
  End = 'end',
  Center = 'center',
}

export default Vue.component('ThSort', {
  functional: true,
  props: {
    value: {
      type: String,
      default: '',
    },
    order: {
      type: Object as PropType<UseOrder>,
      required: true,
    },
    align: {
      type: String as PropType<Align>,
      default: 'start',
    },
  },
  render(h, { props, data, slots }) {
    const content = [h('span', slots().default)]

    const icon = h(
      VIcon,
      {
        props: { size: 18 },
        class: 'v-data-table-header__icon',
      },
      'mdi-arrow-up',
    )

    let textAlign: string

    switch (props.align) {
      case Align.Start:
        textAlign = 'text-left'
        content.push(icon)
        break
      case Align.End:
        textAlign = 'text-right'
        content.unshift(icon)
        break
      case Align.Center:
        textAlign = 'text-center'
        content.push(icon)
        break
      default:
        throw new Error(`unknown ${props.align}`)
    }

    return h(
      'th',
      {
        ...data,
        class: ['text-no-wrap', textAlign, props.order.thClass(props.value)],
        on: {
          click() {
            props.order.toggle(props.value)
          },
        },
      },
      content,
    )
  },
})
</script>

<style lang="scss" scoped></style>
