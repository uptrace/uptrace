import { min, addMilliseconds, subMilliseconds, differenceInMilliseconds } from 'date-fns'
import { ref, computed, proxyRefs } from '@vue/composition-api'

// Composables
import { useQuery } from '@/use/router'
import { useForceReload } from '@/use/force-reload'

// Utilities
import { formatUTC, parseUTC, toUTC, toLocal, ceilDate, second, minute, day } from '@/util/date'

export type UseDateRange = ReturnType<typeof useDateRange>

export function useDateRange() {
  const lt = ref<Date>()
  const isNow = ref(false)
  const duration = ref(0)

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
      const [hours, minutes] = s.split(':')
      dt.setHours(parseInt(hours, 10))
      dt.setMinutes(parseInt(minutes, 10))
      changeGTE(dt)
    },
  })

  function updateNow(force = false): boolean {
    if (!force && !isNow.value) {
      return false
    }

    isNow.value = true
    const nowVal = ceilDate(new Date(), minute)

    if (nowVal <= lt.value!) {
      return false
    }
    lt.value = nowVal

    if (updateNowTimer) {
      clearTimeout(updateNowTimer)
    }
    updateNowTimer = setTimeout(updateNow, 5 * minute)

    return true
  }

  function resetNow() {
    if (updateNowTimer) {
      clearTimeout(updateNowTimer)
      updateNowTimer = null
    }
    isNow.value = false
  }

  function reload() {
    updateNow(true)
    forceReload()
  }

  function reset() {
    resetNow()
    lt.value = undefined
    duration.value = 0
  }

  function change(gteVal: Date, ltVal: Date) {
    const durVal = ltVal.getTime() - gteVal.getTime()

    if (lt.value && lt.value.getTime() === ltVal.getTime() && duration.value === durVal) {
      return
    }

    resetNow()
    lt.value = ltVal
    duration.value = durVal
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

  function changeWithin(dt: Date | string, ms = 0) {
    if (typeof dt === 'string') {
      dt = new Date(dt)
    }
    if (ms) {
      duration.value = ms
    }

    dt = addMilliseconds(dt, duration.value / 2)
    const now = ceilDate(new Date(), minute)
    dt = min([dt, now])
    changeLT(dt)
  }

  //------------------------------------------------------------------------------

  function changeGTE(dt: Date) {
    changeLT(addMilliseconds(dt, duration.value))
  }

  function changeLT(dt: Date) {
    resetNow()
    lt.value = dt
  }

  //------------------------------------------------------------------------------

  const hasPrevPeriod = computed((): boolean => {
    if (!isValid.value) {
      return false
    }
    const ms = differenceInMilliseconds(new Date(), gte.value!)
    return ms < 30 * day
  })

  function prevPeriod() {
    resetNow()
    lt.value = subMilliseconds(lt.value!, duration.value)
  }

  const hasNextPeriod = computed((): boolean => {
    if (!isValid.value || isNow.value) {
      return false
    }
    const ms = differenceInMilliseconds(new Date(), lt.value!)
    return ms > 15 * minute
  })

  function nextPeriod() {
    const ltVal = addMilliseconds(lt.value!, duration.value)
    const nowVal = new Date()
    changeLT(ltVal <= nowVal ? ltVal : nowVal)
  }

  //------------------------------------------------------------------------------

  function queryParams() {
    if (!isValid.value) {
      return {}
    }
    if (isNow.value) {
      return {
        ['time_dur']: duration.value / second,
      }
    }

    return {
      ['time_gte']: formatUTC(gte.value!),
      ['time_dur']: duration.value / second,
    }
  }

  function parseQueryParams(params: Record<string, any>) {
    const dur = params['time_dur']
    const gte = params['time_gte']
    if (!dur) {
      return
    }

    duration.value = parseInt(dur, 10) * second

    if (typeof gte === 'string') {
      changeGTE(parseUTC(gte))
    } else {
      updateNow(true)
    }
  }

  function axiosParams() {
    if (!isValid.value) {
      return {
        time_gte: undefined,
        time_lt: undefined,
      }
    }

    let gteVal = gte.value!
    let ltVal = lt.value!

    const params: Record<string, any> = {
      ...forceReloadParams.value,
      time_gte: gteVal.toISOString(),
      time_lt: ltVal.toISOString(),
    }

    return params
  }

  function lokiParams() {
    if (!isValid.value) {
      return {
        start: undefined,
        end: undefined,
      }
    }

    return {
      start: gte.value!.getTime() * 1e6,
      end: lt.value!.getTime() * 1e6,
    }
  }

  function syncQuery() {
    useQuery().sync({
      fromQuery(q) {
        parseQueryParams(q)
      },
      toQuery() {
        return queryParams()
      },
    })
  }

  return proxyRefs({
    gte,
    lt,

    isValid,
    isNow,
    duration,

    datePicker,
    timePicker,

    updateNow,
    reload,
    forceReload,

    reset,
    change,
    changeDuration,
    changeWithin,

    changeGTE,
    changeLT,

    hasPrevPeriod,
    prevPeriod,
    hasNextPeriod,
    nextPeriod,

    queryParams,
    axiosParams,
    lokiParams,
    syncQuery,
  })
}
