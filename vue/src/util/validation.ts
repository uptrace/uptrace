export const requiredRule = (v: string) => (v && v.length != 0) || 'Field is required'

export const emailRule = (v: string) => isEmail(v) || 'E-mail must be valid'

export function isEmail(v: string): boolean {
  return /.+@.+/.test(v)
}
