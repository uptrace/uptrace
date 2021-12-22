<script lang="ts">
import Vue from 'vue'
import { PropType } from '@vue/composition-api'
import { VBtn, VIcon } from 'vuetify/lib'

// Composables
import { UsePager } from '@/use/pager'

export default Vue.component('XPager', {
  functional: true,
  props: {
    pager: {
      type: Object as PropType<UsePager>,
      required: true,
    },
    withoutPrevNext: {
      type: Boolean,
      default: false,
    },
  },
  render(h, { props, data }) {
    const children: any[] = [
      h(
        'span',
        { class: 'text-body-2' },
        `${props.pager.pos.start + 1} - ${props.pager.pos.end} of ` + `${props.pager.numItem}`,
      ),
    ]

    if (!props.withoutPrevNext) {
      children.push(
        h(
          VBtn,
          {
            class: { 'ml-2': true },
            props: {
              icon: true,
              disabled: !props.pager.hasPrev,
            },
            on: {
              click: props.pager.prev,
            },
          },
          [h(VIcon, 'mdi-chevron-left')],
        ),
      )
      children.push(
        h(
          VBtn,
          {
            props: {
              icon: true,
              disabled: !props.pager.hasNext,
            },
            on: {
              click: props.pager.next,
            },
          },
          [h(VIcon, 'mdi-chevron-right')],
        ),
      )
    }

    return h(
      'span',
      {
        ...data,
      },
      children,
    )
  },
})
</script>

<style lang="scss" scoped></style>
