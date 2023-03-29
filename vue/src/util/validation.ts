export const requiredRule = (v: string) => (v && v.length != 0) || 'Field is required'

export const emailRule = (v: string) => isEmail(v) || 'E-mail must be valid'

const emailRe = /^[^\s<>]+@[^\s<>]+\.[^\s<>]+$/

export function isEmail(v: string): boolean {
  return emailRe.test(v)
}

export function optionalRule(v: string): boolean {
  return true
}
