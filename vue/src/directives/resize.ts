import { DirectiveOptions } from 'vue'
// eslint-disable-next-line @typescript-eslint/no-var-requires
const elementResizeDetectorMaker = require('element-resize-detector')

const erd = elementResizeDetectorMaker({
  strategy: 'scroll',
})

const directive: DirectiveOptions = {
  bind(el, binding, vnode) {
    let options = {}

    if (typeof binding.value === 'boolean') {
      if (!binding.value) {
        return
      }
    } else if (binding.value) {
      options = binding.value
    }

    erd.listenTo(options, el, (element: HTMLElement) => {
      const event = { width: element.offsetWidth, height: element.offsetHeight }

      if (vnode.componentInstance) {
        vnode.componentInstance.$emit('resize', { detail: event })
      } else if (vnode.elm) {
        vnode.elm.dispatchEvent(new CustomEvent('resize', { detail: event }))
      }
    })
  },

  unbind(el) {
    erd.uninstall(el)
  },
}

export default directive
