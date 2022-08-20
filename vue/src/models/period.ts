import { minute, hour, day } from '@/util/fmt/date'

export interface Period {
  text: string
  ms: number
}

export function periodsForDays(days: number): Period[] {
  const periods = []

  for (let n of [15, 30]) {
    periods.push({
      text: `${n} ${n === 1 ? 'minute' : 'minutes'}`,
      ms: n * minute,
    })
  }

  for (let n of [1, 3, 6, 12, 24]) {
    periods.push({
      text: `${n} ${n === 1 ? 'hour' : 'hours'}`,
      ms: n * hour,
    })
  }

  for (let n of [3, 7, 10, 14, 30]) {
    if (n <= days) {
      periods.push({ text: `${n} days`, ms: n * day })
    }
  }

  return periods
}
