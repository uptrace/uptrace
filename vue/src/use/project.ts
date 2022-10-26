import { computed, proxyRefs } from 'vue'

import { useRouter } from '@/use/router'
import { useWatchAxios } from '@/use/watch-axios'

export interface ConnDetails {
  endpoint: string
  dsn: string
}

export function useProject() {
  const { route } = useRouter()

  const { data } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/projects/${projectId}`,
    }
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
    return data.value?.pinnedAttrs ?? []
  })

  return proxyRefs({ grpc, http, pinnedAttrs })
}
