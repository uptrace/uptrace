export const requiredRule = (v: string) => (v && v.length != 0) || 'Field is required'

export const emailRule = (v: string) => isEmail(v) || 'E-mail must be valid'

const emailRe = /^[^\s<>]+@[^\s<>]+\.[^\s<>]+$/

export function isEmail(v: string): boolean {
  return emailRe.test(v)
}

export function optionalRule(v: string): boolean {
  return true
}

export function minMaxStringLengthRule(min: number, max: number) {
  return (s: string) => {
    const length = s.length
    if (length < min) {
      return `Must be at least ${min} characters long`
    }
    if (length > max) {
      return `Must be no more than ${max} characters long`
    }
    return true
  }
}
