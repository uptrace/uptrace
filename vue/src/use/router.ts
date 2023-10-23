import querystring from 'querystring'
import { omit, debounce } from 'lodash-es'
import { Route } from 'vue-router'
import { computed, watch, onMounted, onBeforeUnmount } from 'vue'

import router from '@/router'

export function useRouter() {
  const route = computed((): Route => {
    return router.app.$root.$route
  })

  return { router, route }
}

export function useRouterOnly() {
  return router
}

export function useRoute() {
  const route = computed((): Route => {
    return router.app.$root.$route
  })
  return route
}

//------------------------------------------------------------------------------

interface QueryHook {
  toQuery(): Record<string, any> | undefined
  fromQuery(query: Values): void
}

let useSyncQueryParamsCount = 0

export function useSyncQueryParams(hook: QueryHook) {
  const { router, route } = useRouter()

  onMounted(() => {
    useSyncQueryParamsCount++
    if (useSyncQueryParamsCount > 1) {
      console?.warn('useSyncQueryParams can not be called more than once on a page')
    }
  })
  onBeforeUnmount(() => {
    useSyncQueryParamsCount--
  })

  // Whether to push a new entry into the history stack or replace an existing one.
  let usePushRoute = false

  const usePushRouteDebounced = debounce(() => {
    usePushRoute = true
  }, 2000)

  let ignoreNextRoute: Route | undefined

  const updateQueryDebounced = debounce(
    (route: Route, query: Record<string, any> | undefined): void => {
      if (query === undefined) {
        return
      }
      for (let key in query) {
        if (query[key] === undefined) {
          return
        }
      }

      if (queryEqual(query, route.query)) {
        // Nothing to do.
        usePushRouteDebounced()
        return
      }

      const savedRoute = ignoreNextRoute
      function onError() {
        ignoreNextRoute = savedRoute
      }

      ignoreNextRoute = omit(route, 'matched') as Route
      ignoreNextRoute.query = query

      if (usePushRoute) {
        router.push({ query, hash: route.hash }).catch(onError)
      } else {
        router.replace({ query, hash: route.hash }).catch(onError)
      }
      usePushRouteDebounced()
    },
    50,
  )

  // Parse query params whenever the route is changed.
  watch(
    route,
    (route) => {
      if (!route.matched.length) {
        return
      }
      if (ignoreNextRoute && routeEqual(route, ignoreNextRoute)) {
        return
      }

      hook.fromQuery(new Values(route.query))
      usePushRoute = false

      updateQueryDebounced(route, hook.toQuery())
    },
    { immediate: true, flush: 'post' },
  )

  // Update query params in the current route whenever query is changed.
  watch(
    () => hook.toQuery(),
    (query) => {
      updateQueryDebounced(route.value, query)
    },
    { immediate: true, flush: 'post' },
  )
}

function routeEqual(r1: Route, r2: Route): boolean {
  return r1.path === r2.path && paramsEqual(r1.params, r2.params) && queryEqual(r1.query, r2.query)
}

function paramsEqual(p1: Record<string, any>, p2: Record<string, any>): boolean {
  const k1 = Object.keys(p1).sort()
  const k2 = Object.keys(p2).sort()
  // JSON omits undefined: {tab: undefined}.
  return JSON.stringify(p1, k1) === JSON.stringify(p2, k2)
}

function queryEqual(
  q1: Record<string, any> | undefined,
  q2: Record<string, any> | undefined,
): boolean {
  // handles int/string: 1 !== "1"
  return querystring.stringify(q1) === querystring.stringify(q2)
}

//------------------------------------------------------------------------------

export class Values {
  kvs: Record<string, string[]>

  constructor(kvs: Record<string, any>) {
    this.kvs = normQueryParams(kvs)
  }

  setDefault(key: string, value: any) {
    if (!this.has(key)) {
      this.set(key, value)
    }
  }

  set(key: string, value: any) {
    this.kvs[key] = normQueryParam(value)
  }

  empty(): boolean {
    return Object.keys(this.kvs).length === 0
  }

  has(key: string): boolean {
    return key in this.kvs
  }

  string(key: string, defValue = ''): string {
    const value = this.kvs[key]
    if (value) {
      return value[0]
    }
    return defValue
  }

  boolean(key: string, defValue = false): boolean {
    return ['true', '1'].includes(this.string(key, normQueryValue(defValue)))
  }

  int(key: string, defValue = 0): number {
    return parseInt(this.string(key), 10) || defValue
  }

  array(key: string, defValue = []): string[] {
    return this.kvs[key] ?? defValue
  }

  forEach(fn: (key: string, value: string[]) => void) {
    for (let key in this.kvs) {
      fn(key, this.kvs[key])
    }
  }
}

function normQueryParams(src: Record<string, any>): Record<string, string[]> {
  const dest: Record<string, string[]> = {}
  for (let key in src) {
    dest[key] = normQueryParam(src[key])
  }
  return dest
}

function normQueryParam(value: any): string[] {
  if (Array.isArray(value)) {
    return value.map((el) => normQueryValue(el))
  }
  return [normQueryValue(value)]
}

function normQueryValue(value: any): string {
  if (value === null || value === undefined) {
    return ''
  }

  switch (typeof value) {
    case 'string':
      return value
    case 'number':
      return String(value)
    case 'boolean':
      return String(value)
  }

  return String(value)
}
