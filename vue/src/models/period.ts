import { minute, hour, day } from '@/util/fmt/date'

export interface Period {
  text: string
  milliseconds: number
}

export function periodsForDays(days: number): Period[] {
  const periods = []

  for (let n of [15, 30]) {
    periods.push({
      text: `${n} ${n === 1 ? 'minute' : 'minutes'}`,
      milliseconds: n * minute,
    })
  }

  for (let n of [1, 3, 6, 12, 24]) {
    periods.push({
      text: `${n} ${n === 1 ? 'hour' : 'hours'}`,
      milliseconds: n * hour,
    })
  }

  for (let n of [3, 7, 10, 14, 30]) {
    if (n <= days) {
      periods.push({ text: `${n} days`, milliseconds: n * day })
    }
  }

  return periods
}
