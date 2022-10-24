import { omit, cloneDeep, debounce } from 'lodash-es'
import { Route } from 'vue-router'
import { shallowRef, computed, watch, onBeforeMount, onBeforeUnmount } from 'vue'

import router from '@/router'
import { defineStore } from '@/use/store'

export function useRouter() {
  const route = computed((): Route => {
    return router.app.$root.$route
  })

  return { router, route }
}

export function useRoute() {
  const route = computed((): Route => {
    return router.app.$root.$route
  })
  return route
}

//------------------------------------------------------------------------------

export type Query = Record<string, any>

interface QueryItem {
  toQuery?(): Query | undefined
  fromQuery?(query: Query): void
}

type OnRouteUpdatedHook = (route: Route) => void

export const useRouteQuery = defineStore(() => {
  const { router, route } = useRouter()

  let isFreshRoute = false
  const lastFreshRoute = shallowRef<Route>()

  const routeUpdatedHooks = shallowRef<OnRouteUpdatedHook[]>([])
  const items = shallowRef<QueryItem[]>([])

  const query = computed((): Query | undefined => {
    let query: Query = {}

    for (let key in route.value.query) {
      if (key.startsWith('utm_')) {
        query[key] = route.value.query[key]
      }
    }

    for (let item of items.value) {
      if (!item.toQuery) {
        continue
      }

      const q = item.toQuery()
      if (!q) {
        return undefined
      }
      Object.assign(query, cloneDeep(q))
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

    const onSuccess = () => {
      ignoreNext = omit(route, 'matched') as Route
      ignoreNext.query = query
    }

    if (isFreshRoute) {
      isFreshRoute = false
      router
        .replace({ query, hash: route.hash })
        .then(onSuccess)
        .catch(() => {})
    } else {
      router
        .push({ query, hash: route.hash })
        .then(onSuccess)
        .catch(() => {})
    }
  }, 100)

  function sync(item: QueryItem) {
    //TODO: vue3
    items.value.push(item)
    // eslint-disable-next-line no-self-assign
    items.value = items.value.slice()

    function remove() {
      const idx = items.value.findIndex((v) => v === item)
      if (idx >= 0) {
        items.value.splice(idx, 1)
        // eslint-disable-next-line no-self-assign
        items.value = items.value.slice()
      } else {
        // eslint-disable-next-line no-console
        console.error("can't find the hook")
      }
    }

    onBeforeUnmount(remove)

    if (item.fromQuery) {
      onBeforeMount(() => {
        if (lastFreshRoute.value) {
          item.fromQuery!(lastFreshRoute.value.query)
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
      } else {
        // eslint-disable-next-line no-console
        console.error("can't find the hook")
      }
    })

    onBeforeMount(() => {
      if (lastFreshRoute.value) {
        hook(lastFreshRoute.value)
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

      lastFreshRoute.value = route
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

      isFreshRoute = true
      updateQueryDebounced(route, query.value)
      ignoreNext = omit(route, 'matched') as Route
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

  return { route: lastFreshRoute, onRouteUpdated, sync }
})

function routeEqual(r1: Route, r2: Route): boolean {
  return r1.path === r2.path && paramsEqual(r1.params, r2.params) && queryEqual(r1.query, r2.query)
}

function paramsEqual(p1: Record<string, any>, p2: Record<string, any>): boolean {
  const k1 = Object.keys(p1).sort()
  const k2 = Object.keys(p2).sort()
  return JSON.stringify(p1, k1) === JSON.stringify(p2, k2)
}

function queryEqual(q1: Query | undefined, q2: Query | undefined): boolean {
  const p1 = new URLSearchParams(q1)
  const p2 = new URLSearchParams(q2)
  return p1.toString() === p2.toString()
}
