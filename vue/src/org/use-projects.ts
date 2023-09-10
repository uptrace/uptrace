import { computed, proxyRefs, watch } from 'vue'

// Composables
import { useStorage } from '@/use/local-storage'
import { useRouter } from '@/use/router'
import { useWatchAxios } from '@/use/watch-axios'

export interface Project {
  id: number
  name: string
  groupByEnv: boolean
  groupFuncsByService: boolean
  pinnedAttrs: string[]
  token: string
}

export interface ConnDetails {
  endpoint: string
  dsn: string
}

export function useProject() {
  const { route } = useRouter()
  const { item: lastProjectId } = useStorage('useProject:lastProjectId', 0)

  const { data } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/projects/${projectId}`,
    }
  })

  const project = computed((): Project | undefined => {
    return data.value?.project
  })

  const grpc = computed((): ConnDetails => {
    return (
      data.value?.grpc ?? {
        endpoint: 'http://localhost:14317',
        dsn: 'http://project1_secret_token@localhost:14317/1',
      }
    )
  })

  const http = computed((): ConnDetails => {
    return (
      data.value?.http ?? {
        endpoint: 'http://localhost:14318',
        dsn: 'http://project1_secret_token@localhost:14318/1',
      }
    )
  })

  const pinnedAttrs = computed(() => {
    return project.value?.pinnedAttrs ?? []
  })

  watch(project, (project) => {
    if (project) {
      lastProjectId.value = project.id
    }
  })

  return proxyRefs({
    data: project,
    grpc,
    http,
    pinnedAttrs,
    lastProjectId,
  })
}
