interface GlobalModel<A extends unknown[], R> {
  (...args: A): R
  inject(): R | undefined
}

export function useGlobalStore<A extends any[], R extends Record<string, unknown>>(
  stateName: string,
  _create: (...args: A) => R,
): GlobalModel<A, R> {
  let state: R | undefined

  model.inject = _inject

  function create(...args: A): R {
    if (state === undefined) {
      try {
        state = _create(...args)
      } catch (err) {
        // eslint-disable-next-line no-console
        console.error(err)
      }
    }
    return state as R
  }

  function model(...args: A): R {
    return create(...args)
  }

  function _inject(): R | undefined {
    return state
  }

  return model
}
