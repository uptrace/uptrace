import { min, addMilliseconds, subMilliseconds, differenceInMilliseconds } from 'date-fns'
import { shallowRef, computed, proxyRefs, watch, onBeforeUnmount } from 'vue'
import { useSessionStorage } from '@vueuse/core'

// Composables
import { useRoute, Values } from '@/use/router'
import { provideForceReload, injectForceReload } from '@/use/force-reload'

// Misc
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
} from '@/util/fmt/date'

const AUTO_RELOAD_INTERVAL = MINUTE

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

  const defaultPrefix = conf.prefix ?? 'time_'
  const roundUpEnabled = shallowRef(false)

  const lt = shallowRef<Date>()
  const isNow = shallowRef(false)
  const duration = shallowRef(0) // milliseconds

  let updateNowTimer: ReturnType<typeof global.setTimeout> | undefined
  const forceReload = injectForceReload()

  const isValid = computed((): boolean => {
    return Boolean(lt.value) && Boolean(duration.value)
  })

  const gte = computed((): Date | undefined => {
    if (!isValid.value) {
      return
    }
    let gte = subMilliseconds(lt.value!, duration.value)
    //gte = truncDate(gte, durationPeriod(duration.value))
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
      const [hours, minutes] = s.split(':')
      dt.setHours(parseInt(hours, 10))
      dt.setMinutes(parseInt(minutes, 10))
      changeGTE(dt)
    },
  })

  const autoReloadEnabled = useSessionStorage('auto-reload-enabled', false)
  function toggleAutoReload() {
    autoReloadEnabled.value = !autoReloadEnabled.value
    if (autoReloadEnabled.value) {
      updateNow(true)
    }
  }

  onBeforeUnmount(() => {
    if (updateNowTimer) {
      clearTimeout(updateNowTimer)
      updateNowTimer = undefined
    }
  })

  watch(
    () => route.value.name,
    () => {
      updateNow()
    },
    { flush: 'post' },
  )

  function reload() {
    updateNow()
    forceReload.do()
  }

  function reloadNow() {
    updateNow(true)
    forceReload.do()
  }

  function reset() {
    resetUpdateNowTimer()
    lt.value = undefined
    duration.value = 0
  }

  function updateNow(force = false) {
    if (!isNow.value) {
      if (!force) {
        return
      }
      isNow.value = true
    }

    let now = new Date()
    now = roundUpEnabled.value ? ceilDate(now, MINUTE) : truncDate(now, MINUTE)

    if (!force && duration.value >= 6 * HOUR) {
      const diff = differenceInMilliseconds(lt.value!, now)
      if (Math.abs(diff) < 5 * MINUTE) {
        // Don't update the time so the query cache is not invalidated.
        return
      }
    }

    lt.value = now

    if (autoReloadEnabled.value) {
      if (updateNowTimer) {
        clearTimeout(updateNowTimer)
      }
      updateNowTimer = global.setTimeout(updateNow, AUTO_RELOAD_INTERVAL)
    }
  }

  function resetUpdateNowTimer() {
    if (updateNowTimer) {
      clearTimeout(updateNowTimer)
      updateNowTimer = undefined
    }
    isNow.value = false
  }

  function change(gteVal: Date, ltVal: Date) {
    const durVal = ltVal.getTime() - gteVal.getTime()
    lt.value = ltVal
    duration.value = durVal
    resetUpdateNowTimer()
  }

  function changeDuration(ms: number): void {
    // Try to preserve gte value.
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

  function syncWith(other: UseDateRange) {
    lt.value = other.lt
    duration.value = other.duration
    isNow.value = other.isNow
  }

  function includes(dt: Date | string): boolean {
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

    if ((ms === 0 || duration.value === ms) && includes(dt)) {
      // Don't change date range if possible.
      return
    }

    if (ms) {
      duration.value = ms
    }

    dt = addMilliseconds(dt, duration.value / 2)
    const now = ceilDate(new Date(), MINUTE) // always round up
    dt = min([dt, now])
    changeLT(dt)
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
    return true
    // const ms = differenceInMilliseconds(new Date(), gte.value!)
    // return ms < 30 * DAY
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

  function queryParams(prefix = 'time_') {
    if (!isValid.value) {
      return {}
    }

    return {
      [prefix + 'gte']: formatUTC(gte.value!),
      [prefix + 'dur']: duration.value / SECOND,
    }
  }

  function parseQueryParams(queryParams: Values, prefix = 'time_') {
    const within = queryParams.string(prefix + 'within')
    if (within) {
      const dt = parseUTC(within)
      changeAround(dt, HOUR)
      return
    }

    const dur = queryParams.int(prefix + 'dur') * SECOND
    if (!dur) {
      // Preserve the current date range.
      return
    }
    duration.value = dur

    const gte = parseUTC(queryParams.string(prefix + 'gte'))
    if (gte) {
      const lt = addMilliseconds(gte, duration.value)
      const ms = differenceInMilliseconds(lt, new Date())
      if (Math.abs(ms) > 5 * MINUTE) {
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
      // Return undefined to block axios request.
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
      ...forceReload.params,
      [prefix + 'gte']: gteVal.toISOString(),
      [prefix + 'lt']: ltVal.toISOString(),
    }

    return params
  }

  function roundUp() {
    roundUpEnabled.value = true
    updateNow()

    onBeforeUnmount(() => {
      roundUpEnabled.value = false
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
    duration,

    isValid,
    isNow,

    datePicker,
    timePicker,

    reload,
    reloadNow,
    autoReloadEnabled,
    toggleAutoReload,

    reset,
    change,
    changeDuration,
    syncWith,

    includes,
    changeAround,

    changeGTE,
    changeLT,

    hasPrevPeriod,
    prevPeriod,
    hasNextPeriod,
    nextPeriod,

    roundUp,
    toArray,

    axiosParams,
    queryParams,
    parseQueryParams,
  })
}

export function useDateRangeFrom(other: UseDateRange | undefined) {
  provideForceReload()
  const dateRange = useDateRange()
  if (other) {
    dateRange.syncWith(other)
  }
  return dateRange
}
