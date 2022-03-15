// MIT License

// Copyright (c) 2017 Collin Henderson

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { DirectiveOptions } from 'vue'
import { DirectiveBinding } from 'vue/types/options'

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
