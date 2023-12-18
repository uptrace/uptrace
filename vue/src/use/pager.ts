import { shallowRef, computed, watch, proxyRefs } from 'vue'

interface PagePos {
  start: number
  end: number
}

export type UsePager = ReturnType<typeof usePager>

export function usePager(perPageValue = 10) {
  const numItem = shallowRef(0)
  const page = shallowRef(1)
  const perPage = shallowRef(perPageValue)

  const pos = computed((): PagePos => {
    const start = (page.value - 1) * perPage.value
    let end = start + perPage.value
    if (end > numItem.value) {
      end = numItem.value
    }
    return { start, end }
  })

  const numPage = computed(() => {
    const maxNumPage = 1000

    const numPage = Math.ceil(numItem.value / perPage.value)
    if (numPage > maxNumPage) {
      return maxNumPage
    }
    return numPage
  })

  const hasPrev = computed(function () {
    return page.value > 1
  })
  function prev() {
    if (hasPrev.value) {
      page.value--
    }
  }

  const hasNext = computed(function () {
    return page.value < numPage.value
  })
  function next() {
    if (hasNext.value) {
      page.value++
    }
  }

  function reset() {
    page.value = 1
  }

  function axiosParams() {
    return {
      page: page.value,
      limit: perPage.value,
    }
  }

  watch(numPage, (numPage) => {
    if (page.value > numPage) {
      page.value = 1
    }
  })

  return proxyRefs({
    numItem,
    page,
    perPage,

    pos,
    numPage,

    hasPrev,
    prev,
    hasNext,
    next,

    reset,
    axiosParams,
  })
}
