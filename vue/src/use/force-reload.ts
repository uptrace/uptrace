import { ref, computed } from '@vue/composition-api'

import { defineStore } from '@/use/store'

export const useForceReload = defineStore(() => {
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
