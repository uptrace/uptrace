import { computed, proxyRefs } from 'vue'

import router from '@/router'

// Composables
import { useRoute } from '@/use/router'
import { defineStore } from '@/use/store'
import { useAxios } from '@/use/axios'
import { useWatchAxios } from '@/use/watch-axios'
import { Project } from '@/org/use-projects'

export interface User {
  name: string
  email: string
  avatar: string
}

export const useUser = defineStore(() => {
  const route = useRoute()
  const { loading, data, request } = useAxios()

  const user = computed((): User | undefined => {
    return data.value?.user
  })

  const isAuth = computed((): boolean => {
    return user.value !== undefined
  })

  const projects = computed((): Project[] => {
    return data.value?.projects ?? []
  })

  const activeProject = computed((): Project | undefined => {
    const projectId = parseInt(route.value.params.projectId)
    if (!projectId) {
      return
    }
    for (let p of projects.value) {
      if (p.id === projectId) {
        return p
      }
    }
    return undefined
  })

  let req: Promise<any>

  getOrLoad()

  function reload() {
    req = request({ url: '/api/v1/users/current' })
    return req
  }

  async function getOrLoad() {
    if (!req) {
      reload()
    }
    return await req
  }

  function logout() {
    return request({ method: 'POST', url: '/api/v1/users/logout' }).then(() => {
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
    activeProject,

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
    return { url: '/api/v1/sso/methods' }
  })

  const methods = computed((): SsoMethod[] => {
    return data.value?.methods ?? []
  })

  return proxyRefs({ loading, methods })
}
