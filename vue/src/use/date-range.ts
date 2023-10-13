import { min, addMilliseconds, subMilliseconds, differenceInMilliseconds } from 'date-fns'
import { shallowRef, computed, proxyRefs, watch, onBeforeUnmount, getCurrentInstance } from 'vue'

// Composables
import { useRoute, useRouteQuery } from '@/use/router'
import { useForceReload } from '@/use/force-reload'

// Utilities
import {
  formatUTC,
  parseUTC,
  toUTC,
  toLocal,
  ceilDate,
  truncDate,
  SECOND,
  MINUTE,
  HOUR,
  DAY,
} from '@/util/fmt/date'

const UPDATE_NOW_TIMER_DELAY = 5 * MINUTE

export type UseDateRange = ReturnType<typeof useDateRange>

interface Config {
  prefix?: string
}

interface ParamsConfig {
  prefix?: string
  optional?: boolean
  offset?: number
}

export function useDateRange(conf: Config = {}) {
  const route = useRoute()
  let _roundUp = false
  const defaultPrefix = conf.prefix ?? 'time_'

  const lt = shallowRef<Date>()
  const isNow = shallowRef(false)
  const duration = shallowRef(0)

  let updateNowTimer: ReturnType<typeof setTimeout> | null
  const { forceReload, forceReloadParams } = useForceReload()

  const isValid = computed((): boolean => {
    return Boolean(lt.value) && Boolean(duration.value)
  })

  const gte = computed((): Date | undefined => {
    if (!isValid.value) {
      return
    }
    let gte = subMilliseconds(lt.value!, duration.value)
    return gte
  })

  // For v-date-picker.
  const datePicker = computed({
    get(): string {
      return toLocal(gte.value!).toISOString().substr(0, 10)
    },
    set(s: string) {
      const dt = toUTC(new Date(s))
      changeGTE(dt)
    },
  })

  // For v-time-picker.
  const timePicker = computed({
    get(): string {
      const dt = gte.value!
      return `${dt.getHours()}:${dt.getMinutes()}`
    },
    set(s: string) {
      const dt = new Date(gte.value!.getTime())
      const [HOURs, MINUTEs] = s.split(':')
      dt.setHours(parseInt(HOURs, 10))
      dt.setMinutes(parseInt(MINUTEs, 10))
      changeGTE(dt)
    },
  })

  watch(
    () => route.value.name,
    () => {
      updateNow()
    },
    { flush: 'post' },
  )

  if (getCurrentInstance()) {
    onBeforeUnmount(() => {
      if (updateNowTimer) {
        clearTimeout(updateNowTimer)
        updateNowTimer = null
      }
    })
  }

  function updateNow(force = false): boolean {
    if (force) {
      isNow.value = true
    } else if (!isNow.value) {
      return false
    }

    const now = _roundUp ? ceilDate(new Date(), MINUTE) : truncDate(new Date(), MINUTE)
    lt.value = now

    if (updateNowTimer) {
      clearTimeout(updateNowTimer)
    }
    updateNowTimer = setTimeout(updateNow, UPDATE_NOW_TIMER_DELAY)

    return true
  }

  function resetUpdateNowTimer() {
    if (updateNowTimer) {
      clearTimeout(updateNowTimer)
      updateNowTimer = null
    }
    isNow.value = false
  }

  function reload() {
    updateNow()
    forceReload()
  }

  function reloadNow() {
    updateNow(true)
    forceReload()
  }

  function reset() {
    resetUpdateNowTimer()
    lt.value = undefined
    duration.value = 0
  }

  function change(gteVal: Date, ltVal: Date) {
    const durVal = ltVal.getTime() - gteVal.getTime()
    lt.value = ltVal
    duration.value = durVal
    resetUpdateNowTimer()
  }

  function changeDuration(ms: number): void {
    if (lt.value && !isNow.value) {
      const newLT = addMilliseconds(gte.value!, ms)
      const now = new Date()
      if (newLT < now) {
        duration.value = ms
        lt.value = newLT
        return
      }
    }

    duration.value = ms
    updateNow(true)
  }

  function contains(dt: Date | string): boolean {
    if (!isValid.value) {
      return false
    }
    if (typeof dt === 'string') {
      dt = new Date(dt)
    }
    return dt >= gte.value! && dt < lt.value!
  }

  function changeAround(dt: Date | string, ms = 0) {
    if (typeof dt === 'string') {
      dt = new Date(dt)
    }
    if (ms) {
      duration.value = ms
    }

    dt = addMilliseconds(dt, duration.value / 2)
    const now = ceilDate(new Date(), MINUTE)
    dt = min([dt, now])
    changeLT(dt)
  }

  function syncWith(other: UseDateRange) {
    lt.value = other.lt
    duration.value = other.duration
    isNow.value = other.isNow
  }

  //------------------------------------------------------------------------------

  function changeGTE(dt: Date) {
    changeLT(addMilliseconds(dt, duration.value))
  }

  function changeLT(dt: Date) {
    resetUpdateNowTimer()
    lt.value = dt
  }

  //------------------------------------------------------------------------------

  const hasPrevPeriod = computed((): boolean => {
    if (!isValid.value) {
      return false
    }
    const ms = differenceInMilliseconds(new Date(), gte.value!)
    return ms < 30 * DAY
  })

  function prevPeriod() {
    resetUpdateNowTimer()
    lt.value = subMilliseconds(lt.value!, duration.value)
  }

  const hasNextPeriod = computed((): boolean => {
    if (!isValid.value || isNow.value) {
      return false
    }
    const ms = differenceInMilliseconds(new Date(), lt.value!)
    return ms > 15 * MINUTE
  })

  function nextPeriod() {
    const ltVal = addMilliseconds(lt.value!, duration.value)
    const nowVal = new Date()
    changeLT(ltVal <= nowVal ? ltVal : nowVal)
  }

  //------------------------------------------------------------------------------

  function syncQueryParams(conf: ParamsConfig = {}) {
    useRouteQuery().sync({
      fromQuery(params) {
        parseQueryParams(params, conf)
      },
      toQuery() {
        return queryParams(conf)
      },
    })
  }

  function queryParams(conf: ParamsConfig = {}) {
    const prefix = conf.prefix ?? defaultPrefix

    if (!isValid.value) {
      return {}
    }

    return {
      [prefix + 'gte']: formatUTC(gte.value!),
      [prefix + 'dur']: duration.value / SECOND,
    }
  }

  function parseQueryParams(params: Record<string, any>, conf: ParamsConfig = {}) {
    if (!Object.keys(params)) {
      return
    }

    const prefix = conf.prefix ?? defaultPrefix

    const within = params[prefix + 'within']
    if (typeof within === 'string') {
      const dt = parseUTC(within)
      changeAround(dt, HOUR)
      return
    }

    const dur = params[prefix + 'dur']
    const gteParam = params[prefix + 'gte']
    if (!dur) {
      return
    }

    duration.value = parseInt(dur, 10) * SECOND

    if (typeof gteParam === 'string') {
      const gte = parseUTC(gteParam)
      const lt = addMilliseconds(gte, duration.value)
      const ms = differenceInMilliseconds(lt, new Date())
      if (Math.abs(ms) > UPDATE_NOW_TIMER_DELAY) {
        changeLT(lt)
        return
      }
    }

    updateNow(true)
  }

  function axiosParams(conf: ParamsConfig = {}) {
    const prefix = conf.prefix ?? defaultPrefix

    if (!isValid.value) {
      if (conf.optional) {
        return {}
      }
      return {
        [prefix + 'gte']: undefined,
        [prefix + 'lt']: undefined,
      }
    }

    let gteVal = gte.value!
    let ltVal = lt.value!

    if (conf.offset) {
      gteVal = addMilliseconds(gteVal, conf.offset)
      ltVal = addMilliseconds(ltVal, conf.offset)
    }

    const params: Record<string, any> = {
      ...forceReloadParams.value,
      [prefix + 'gte']: gteVal.toISOString(),
      [prefix + 'lt']: ltVal.toISOString(),
    }

    return params
  }

  function roundUp() {
    _roundUp = true
    updateNow()

    onBeforeUnmount(() => {
      _roundUp = false
      updateNow()
    })
  }

  function toArray() {
    if (isValid.value) {
      return [gte.value, lt.value]
    }
    return []
  }

  return proxyRefs({
    gte,
    lt,

    isValid,
    isNow,
    duration,

    datePicker,
    timePicker,

    reload,
    reloadNow,

    reset,
    change,
    changeDuration,
    contains,
    changeAround,
    syncWith,

    changeGTE,
    changeLT,

    hasPrevPeriod,
    prevPeriod,
    hasNextPeriod,
    nextPeriod,

    queryParams,
    parseQueryParams,
    syncQueryParams,
    roundUp,

    axiosParams,
    toArray,
  })
}
