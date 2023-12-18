import { shallowRef, computed, proxyRefs } from 'vue'
import { injectLocal, provideLocal } from '@vueuse/core'

export type UseForceReload = ReturnType<typeof useForceReload>

function useForceReload() {
  const loading = shallowRef(false)
  const token = shallowRef(0)

  function forceReload() {
    const slot = Math.trunc(Date.now() / 3000)
    if (slot === token.value) {
      return
    }
    token.value = slot

    loading.value = true
    setTimeout(() => {
      loading.value = false
    }, 1000)
  }

  const params = computed(() => {
    if (token.value) {
      return {
        _force: token.value,
      }
    }
    return {}
  })

  return proxyRefs({ loading, params, do: forceReload })
}

const forceReloadKey = Symbol('force-reload')

export function provideForceReload() {
  provideLocal(forceReloadKey, useForceReload())
}

export function injectForceReload(): UseForceReload {
  return injectLocal(forceReloadKey)!
}
