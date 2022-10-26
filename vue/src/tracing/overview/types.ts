export interface OverviewItem {
  system: string

  spanCount: number
  errorCount: number
  errorPct: number

  p50: number
  p90: number
  p99: number
  max: number
  rate: number

  stats: Stats
}

interface Stats {
  count: number[]
  errorCount: number[]
  p50: number[]
  p90: number[]
  p99: number[]
  time: string[]
}
