import { computed, proxyRefs } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { useAxios } from '@/use/axios'
import { useWatchAxios, AxiosParamsSource } from '@/use/watch-axios'
import { User } from '@/org/use-users'

export interface SavedView {
  id: number

  userId: number
  projectId: number

  name: string
  route: string
  params: Record<string, any>
  query: Record<string, any>
  pinned: boolean

  user?: User
}

export type UseSavedViews = ReturnType<typeof useSavedViews>

export function useSavedViews(paramsSource: AxiosParamsSource) {
  const route = useRoute()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const { projectId } = route.value.params
    return {
      url: `/api/v1/tracing/${projectId}/saved-views`,
      params: paramsSource(),
    }
  })

  const views = computed(() => {
    return data.value?.views ?? []
  })

  return proxyRefs({
    status,
    loading,

    items: views,

    reload,
  })
}

export function useSavedViewManager() {
  const route = useRoute()
  const { loading: pending, request } = useAxios()

  function save(view: Partial<SavedView>) {
    const { projectId } = route.value.params
    const url = `/api/v1/tracing/${projectId}/saved-views`
    return request({ method: 'POST', url, data: view })
  }

  function del(viewId: number) {
    const { projectId } = route.value.params
    const url = `/api/v1/tracing/${projectId}/saved-views/${viewId}`
    return request({ method: 'DELETE', url })
  }

  function pin(viewId: number) {
    const { projectId } = route.value.params
    const url = `/api/v1/tracing/${projectId}/saved-views/${viewId}/pinned`
    return request({ method: 'PUT', url })
  }

  function unpin(viewId: number) {
    const { projectId } = route.value.params
    const url = `/api/v1/tracing/${projectId}/saved-views/${viewId}/unpinned`
    return request({ method: 'PUT', url })
  }

  return proxyRefs({ pending, save, del, pin, unpin })
}
