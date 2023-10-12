import { computed, proxyRefs, provide, inject, watch, ComputedRef } from 'vue'

// Composables
import { useRoute } from '@/use/router'
import { usePager } from '@/use/pager'
import { useAxios } from '@/use/axios'
import { useWatchAxios, AxiosParamsSource } from '@/use/watch-axios'

export interface Annotation {
  id: number
  projectId: number

  name: string
  description: string
  color: string
  attrs: Record<string, string>
  createdAt: string
}

export function emptyAnnotation(): Annotation {
  return {
    id: 0,
    projectId: 0,

    name: '',
    description: '',
    color: '#4CAF50',
    attrs: {},
    createdAt: '',
  }
}

const injectionKey = Symbol('annotations')

export interface Attr {
  key: string
  value: string
}

export function useAnnotations(axiosParamsSource: AxiosParamsSource) {
  const route = useRoute()
  const pager = usePager()

  const { status, loading, data, reload } = useWatchAxios(() => {
    const projectId = route.value.params.projectId
    const req = {
      url: `/api/v1/projects/${projectId}/annotations`,
      params: {
        ...axiosParamsSource(),
        ...pager.axiosParams(),
      },
    }
    return req
  })

  const annotations = computed((): Annotation[] => {
    return data.value?.annotations ?? []
  })

  const count = computed(() => {
    return data.value?.count ?? 0
  })

  provide(injectionKey, annotations)

  watch(count, (count) => {
    pager.numItem = count
  })

  return proxyRefs({
    status,
    loading,
    reload,

    items: annotations,
    pager,
    count,
  })
}

export function useAnnotation() {
  const route = useRoute()

  const { loading, data, reload } = useWatchAxios(() => {
    const { projectId, annotationId } = route.value.params
    const url = `/api/v1/projects/${projectId}/annotations/${annotationId}`
    return { url }
  })

  const annotation = computed((): Annotation | undefined => {
    return data.value?.annotation
  })

  return proxyRefs({ loading, data: annotation, reload })
}

export function useAnnotationManager() {
  const route = useRoute()

  const { loading: pending, request } = useAxios()

  function create(annotation: Partial<Annotation>) {
    const projectId = route.value.params.projectId
    const url = `/api/v1/projects/${projectId}/annotations`
    return request({ method: 'POST', url, data: annotation })
  }

  function update(annotation: Partial<Annotation>) {
    const projectId = route.value.params.projectId
    const url = `/api/v1/projects/${projectId}/annotations/${annotation.id}`
    return request({ method: 'PUT', url, data: annotation })
  }

  function del(annotation: Partial<Annotation>) {
    const projectId = route.value.params.projectId
    const url = `/api/v1/projects/${projectId}/annotations/${annotation.id}`
    return request({ method: 'DELETE', url })
  }

  return proxyRefs({
    pending,

    create,
    update,
    del,
  })
}

export function injectAnnotations() {
  return inject<ComputedRef<Annotation[]>>(
    injectionKey,
    computed(() => {
      return []
    }),
  )
}
