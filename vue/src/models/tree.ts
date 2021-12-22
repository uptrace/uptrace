export interface Tree<T> {
  parent: T | null
  children: T[] | null
}

type Func<T> = (child: T, parent: T | null) => boolean

export function walkTree<T extends Tree<T>>(root: T, fn: Func<T>) {
  if (fn(root, null) === false) {
    return
  }
  _walk<T>(root, fn)
}

function _walk<T extends Tree<T>>(parent: T, fn: Func<T>) {
  if (!parent.children) {
    return true
  }

  for (let child of parent.children) {
    if (fn(child, parent) === false) {
      return false
    }
    if (_walk(child, fn) === false) {
      return false
    }
  }

  return true
}
