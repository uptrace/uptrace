import { reactive } from 'vue'

export interface AttrMatcher {
  attr: string
  op: AttrMatcherOp
  value: string
}

export enum AttrMatcherOp {
  Equal = '=',
  NotEqual = '!=',
}

export function emptyAttrMatcher(): AttrMatcher {
  return reactive({ attr: '', op: AttrMatcherOp.Equal, value: '' })
}
