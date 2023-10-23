import { debounce } from 'lodash-es'
import type { Ref } from 'vue-demi'
import { ref, watch } from 'vue-demi'

interface DebouncedRef<T> extends Ref<T> {
  flush: () => void
  cancel: () => void
}
/**
 * Debounce updates of a ref.
 *
 * @return A new debounced ref.
 */
export function refDebounced<T>(value: Ref<T>, ms = 200): Readonly<DebouncedRef<T>> {
  const debounced = ref(value.value as T) as Ref<T>

  const updater = debounce(() => {
    debounced.value = value.value
  }, ms)

  Object.defineProperty(debounced, 'flush', {
    value() {
      updater()
      updater.flush()
    },
    enumerable: false,
  })
  Object.defineProperty(debounced, 'cancel', { value: updater.cancel, enumerable: false })

  watch(value, () => updater())

  return debounced as DebouncedRef<T>
}
