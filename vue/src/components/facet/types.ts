export enum Category {
  All = 'all',
  Pinned = 'pinned',
  Found = 'found',
  Other = 'other',
}

export interface Item {
  value: string
  text?: string
  count: number
}

export interface Filter {
  attr: string
  op: string
  value: string[]
}
