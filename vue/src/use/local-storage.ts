import { computed, ref, watch, Ref } from 'vue'

export function setItem(key: string, value: any) {
  localStorage.setItem(key, JSON.stringify(value))
}

export function getItem(key: string): any {
  const value = localStorage.getItem(key)
  if (value === null) {
    return null
  }
  try {
    return JSON.parse(value)
  } catch {
    return null
  }
}

export function useStorage<T>(key: string | Ref<string>, defValue: T | null = null) {
  const keyRef = ref(key)
  const valueRef = ref()

  const item = computed({
    get(): T {
      return valueRef.value
    },
    set(value: T) {
      setItem(keyRef.value, value)
      valueRef.value = value
    },
  })

  watch(
    keyRef,
    (key) => {
      let value = getItem(key)
      if (value === null) {
        value = defValue
      }
      valueRef.value = value
    },
    { immediate: true, flush: 'sync' },
  )

  return { item }
}
