import { omit, cloneDeep, debounce } from 'lodash'
import { Route } from 'vue-router'
import {
  getCurrentInstance,
  shallowRef,
  computed,
  watch,
  onBeforeMount,
  onBeforeUnmount,
} from 'vue'

import { defineStore } from '@/use/store'

export function useRouter() {
  const vm = getCurrentInstance()
  if (!vm) {
    throw new Error('useRouter is called outside of setup')
  }
  const router = vm.proxy.$router

  const route = computed((): Route => {
    return vm.proxy.$route
  })

  return { router, route }
}

//------------------------------------------------------------------------------

export type Query = Record<string, any>

interface QueryItem {
  toQuery?(): Query | undefined
  fromQuery?(query: Query): void
}

type OnRouteUpdatedHook = (route: Route) => void

export const useQuery = defineStore(() => {
  const { router, route } = useRouter()
  const lastRoute = shallowRef<Route>()

  const routeUpdatedHooks = shallowRef<OnRouteUpdatedHook[]>([])
  const items = shallowRef<QueryItem[]>([])

  const query = computed((): Query | undefined => {
    if (!route.value.matched.length) {
      return
    }

    let query: Query = {}

    for (let item of items.value) {
      if (!item.toQuery) {
        continue
      }

      const q = item.toQuery()
      if (q) {
        Object.assign(query, cloneDeep(q))
      }
    }

    return query
  })

  let ignoreNext: Route | undefined

  const updateQueryDebounced = debounce((route: Route, query: Query | undefined): void => {
    if (query === undefined) {
      return
    }

    for (let k in query) {
      if (query[k] === undefined) {
        return
      }
    }

    ignoreNext = omit(route, 'matched') as Route
    ignoreNext.query = query

    router.replace({ query, hash: route.hash }).catch(() => {})
  }, 10)

  function sync(item: QueryItem) {
    items.value.push(item)
    // eslint-disable-next-line no-self-assign
    items.value = items.value.slice()

    function remove() {
      const idx = items.value.findIndex((v) => v === item)
      if (idx >= 0) {
        items.value.splice(idx, 1)
        // eslint-disable-next-line no-self-assign
        items.value = items.value.slice()
      }
    }

    onBeforeUnmount(remove)

    if (item.fromQuery) {
      onBeforeMount(() => {
        if (lastRoute.value && item.fromQuery) {
          item.fromQuery(lastRoute.value.query)
        }
      })
    }

    return remove
  }

  function onRouteUpdated(hook: OnRouteUpdatedHook) {
    routeUpdatedHooks.value.push(hook)
    // eslint-disable-next-line no-self-assign
    routeUpdatedHooks.value = routeUpdatedHooks.value.slice()

    onBeforeUnmount(() => {
      const idx = routeUpdatedHooks.value.findIndex((v) => v === hook)
      if (idx >= 0) {
        routeUpdatedHooks.value.splice(idx, 1)
        // eslint-disable-next-line no-self-assign
        routeUpdatedHooks.value = routeUpdatedHooks.value.slice()
      }
    })

    onBeforeMount(() => {
      if (lastRoute.value) {
        hook(lastRoute.value)
      }
    })
  }

  watch(
    route,
    (route) => {
      if (!route.matched.length) {
        return
      }
      if (ignoreNext && routeEqual(route, ignoreNext)) {
        return
      }

      for (let item of items.value) {
        if (item.fromQuery) {
          item.fromQuery(route.query)
        }
      }

      for (let hook of routeUpdatedHooks.value) {
        hook(route)
      }

      lastRoute.value = route
    },
    { immediate: true, flush: 'sync' },
  )

  watch(
    route,
    (route) => {
      if (!route.matched.length) {
        return
      }
      if (ignoreNext && routeEqual(route, ignoreNext)) {
        return
      }
      updateQueryDebounced(route, query.value)
    },
    { flush: 'post' },
  )

  watch(
    query,
    (query, oldQuery) => {
      if (!queryEqual(query, oldQuery)) {
        updateQueryDebounced(route.value, query)
      }
    },
    { immediate: true, flush: 'post' },
  )

  return { route: lastRoute, onRouteUpdated, sync }
})

function routeEqual(r1: Route, r2: Route): boolean {
  return r1.path === r2.path && paramsEqual(r1.params, r2.params) && queryEqual(r1.query, r2.query)
}

function paramsEqual(p1: Record<string, any>, p2: Record<string, any>): boolean {
  return JSON.stringify(p1, Object.keys(p1).sort()) === JSON.stringify(p2, Object.keys(p2).sort())
}

function queryEqual(q1: Query | undefined, q2: Query | undefined): boolean {
  const p1 = new URLSearchParams(q1)
  const p2 = new URLSearchParams(q2)
  return p1.toString() === p2.toString()
}
