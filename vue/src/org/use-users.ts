import { computed, proxyRefs, shallowReactive } from 'vue'

import router from '@/router'

// Composables
import { useStorage } from '@/use/local-storage'
import { useRoute } from '@/use/router'
import { useGlobalStore } from '@/use/store'
import { useAxios } from '@/use/axios'
import { useWatchAxios } from '@/use/watch-axios'
import { Project } from '@/org/use-projects'

export interface User {
  id: number
  name: string
  email: string
  avatar: string
}

export const useUser = useGlobalStore('useUser', () => {
  const route = useRoute()
  const { loading, data, request } = useAxios()

  const user = computed((): User => {
    return shallowReactive(data.value?.user ?? { id: 0, name: 'Guest', budget: 0 })
  })

  const isAuth = computed((): boolean => {
    return Boolean(user.value.id)
  })

  const projects = computed((): Project[] => {
    return data.value?.projects ?? []
  })

  const lastProjectId = useStorage(
    computed(() => `last-project-id:${user.value.id}`),
    0,
  )
  const activeProjectId = computed(() => {
    if (route.value.params.projectId) {
      return parseInt(route.value.params.projectId)
    }
    if (!lastProjectId.value) {
      return projects.value[0]?.id
    }
    const found = projects.value.find((p) => p.id === lastProjectId.value)
    if (found) {
      return found.id
    }
    return projects.value[0]?.id
  })

  let req: Promise<any>

  getOrLoad()

  function reload() {
    req = request({ url: '/internal/v1/users/current' })
    return req
  }

  async function getOrLoad() {
    if (!req) {
      reload()
    }
    return await req
  }

  function logout() {
    return request({ method: 'POST', url: '/internal/v1/users/logout' }).then(() => {
      reload().finally(() => {
        redirectToLogin()
      })
    })
  }

  return proxyRefs({
    loading,
    current: user,
    isAuth,
    projects,
    lastProjectId,
    activeProjectId,

    reload,
    getOrLoad,
    logout,
  })
})

export function redirectToLogin() {
  router.push({ name: 'Login' }).catch(() => {})
}

interface SsoMethod {
  name: string
  url: string
}

export function useSso() {
  const { loading, data } = useWatchAxios(() => {
    return { url: '/internal/v1/sso/methods' }
  })

  const methods = computed((): SsoMethod[] => {
    return data.value?.methods ?? []
  })

  return proxyRefs({ loading, methods })
}
