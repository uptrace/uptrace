interface Store<R> {
  (): R
}

export function defineStore<T extends Record<string, unknown>>(
  stateName: string,
  _create: Store<T>,
): Store<T> {
  let store: T | undefined

  return function create(): T {
    if (store === undefined) {
      store = _create()
    }
    return store as T
  }
}
