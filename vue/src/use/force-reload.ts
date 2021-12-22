import { ref, computed } from '@vue/composition-api'

import { useGlobalStore } from '@/use/store'

export const useForceReload = useGlobalStore('useForceReload', () => {
  const token = ref(0)

  function forceReload() {
    token.value = Date.now()
  }

  const forceReloadParams = computed(() => {
    if (token.value) {
      return {
        _force: token.value,
      }
    }
    return {}
  })

  return { forceReload, forceReloadParams }
})
