import { DirectiveOptions } from 'vue'
import { DirectiveBinding } from 'vue/types/options'

// https://github.com/syropian/vue-input-autowidth
const directive: DirectiveOptions = {
  bind(el) {
    const input = findInput(el)
    input.style.boxSizing = 'content-box'
  },
  inserted(el, binding) {
    const input = findInput(el)
    const styles = window.getComputedStyle(input)
    const mirror = document.createElement('div')

    Object.assign(mirror.style, {
      position: 'absolute',
      top: '0',
      left: '0',
      visibility: 'hidden',
      height: '0',
      overflow: 'hidden',
      whiteSpace: 'pre',
      fontSize: styles.fontSize,
      fontFamily: styles.fontFamily,
      fontWeight: styles.fontWeight,
      fontStyle: styles.fontStyle,
      letterSpacing: styles.letterSpacing,
      textTransform: styles.textTransform,
    })

    mirror.setAttribute('aria-hidden', 'true')

    document.body.appendChild(mirror)
    input.mirror = mirror

    checkWidth(input, binding)
    input.addEventListener('input', checkWidth.bind(null, input, binding))
  },
  componentUpdated(el, binding) {
    checkWidth(findInput(el), binding)
  },
  unbind(el, binding) {
    const input = findInput(el)
    document.body.removeChild(input.mirror)
    el.removeEventListener('input', checkWidth.bind(null, input, binding))
  },
}
export default directive

interface MyHTMLInputElement extends HTMLInputElement {
  mirror: HTMLDivElement
}

function findInput(el: HTMLElement): MyHTMLInputElement {
  if (el.tagName.toUpperCase() === 'INPUT') {
    return el as MyHTMLInputElement
  }

  const input = el.querySelector('input')
  if (!input) {
    throw new Error('v-autowidth can only be used on elements with an input')
  }
  return input as MyHTMLInputElement
}

function checkWidth(el: MyHTMLInputElement, binding: DirectiveBinding) {
  const mirror = el.mirror
  const defaults = { maxWidth: 'none', minWidth: 'none', comfortZone: 0 }
  const options = Object.assign({}, defaults, binding.value)

  el.style.maxWidth = options.maxWidth
  el.style.minWidth = options.minWidth

  let val = el.value

  if (!val) {
    val = el.placeholder || ''
  }

  while (mirror.childNodes.length) {
    mirror.removeChild(mirror.childNodes[0])
  }

  mirror.appendChild(document.createTextNode(val))

  let newWidth = mirror.scrollWidth + options.comfortZone + 2

  if (newWidth != el.scrollWidth) {
    el.style.width = `${newWidth}px`
  }
}
