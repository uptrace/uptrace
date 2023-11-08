import { computed, proxyRefs, watch, Ref } from 'vue'

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

export function useProject() {
  const { route } = useRouter()
  const { item: lastProjectId } = useStorage('useProject:lastProjectId', 0)

  const { data } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/internal/v1/projects/${projectId}`,
    }
  })

  const project = computed((): Project | undefined => {
    return data.value?.project
  })

  const dsn = computed((): string => {
    return data.value?.dsn ?? 'http://project1_secret_token@localhost:14318?grpc=14317'
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
    dsn,
    pinnedAttrs,
    lastProjectId,
  })
}

export function useDsn(dsn: Ref<string>) {
  const url = computed(() => {
    return new URL(dsn.value)
  })

  const insecure = computed(() => {
    return url.value.protocol === 'http:'
  })

  const grpcEndpoint = computed(() => {
    switch (url.value.hostname) {
      case 'uptrace.dev':
      case 'api.uptrace.dev':
        return 'https://otlp.uptrace.dev:4317'
      default: {
        const port = url.value.searchParams.get('grpc') ?? 4317
        return `${url.value.protocol}//${url.value.hostname}:${port}`
      }
    }
  })

  const httpEndpoint = computed(() => {
    switch (url.value.hostname) {
      case 'uptrace.dev':
      case 'api.uptrace.dev':
        return 'https://otlp.uptrace.dev'
      default:
        return `${url.value.protocol}//${url.value.host}`
    }
  })

  return proxyRefs({
    original: dsn,
    insecure,
    grpcEndpoint,
    httpEndpoint,
  })
}
