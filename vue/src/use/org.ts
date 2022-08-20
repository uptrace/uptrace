import { computed, proxyRefs } from 'vue'

import router from '@/router'

// Composables
import { defineStore } from '@/use/store'
import { useAxios } from '@/use/axios'

export interface User {
  id: number
  name: string
}

export interface Project {
  id: number
  name: string
}

export const useUser = defineStore(() => {
  const { loading, data, request } = useAxios()

  const user = computed((): User => {
    return data.value?.user ?? { id: 0, name: 'Guest' }
  })

  const isAuth = computed((): boolean => {
    return Boolean(user.value.id)
  })

  const projects = computed((): Project[] => {
    return data.value?.projects ?? []
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

    reload,
    getOrLoad,
    logout,
  })
})

export function redirectToLogin() {
  router.push({ name: 'Login' }).catch(() => {})
}
