import { watch } from 'vue'
import vuetify from '@/plugins/vuetify'
import { useDark, useToggle } from '@vueuse/core'

// Composables
import { defineStore } from '@/use/store'

export const useDarkMode = defineStore(() => {
  const isDark = useDark()
  const toggleDark = useToggle(isDark)

  watch(
    isDark,
    (isDark) => {
      vuetify.framework.theme.dark = isDark
      document.documentElement.style.setProperty('color-scheme', isDark ? 'dark' : 'light')
    },
    { immediate: true },
  )

  return { isDark, toggleDark }
})
