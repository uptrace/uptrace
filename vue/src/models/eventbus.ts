export interface Event {
  scope: Record<string, unknown>
  callback: Callback
}

type Callback = (...args: any[]) => void

export class EventBus {
  events: { [key: string]: Event[] }

  constructor() {
    this.events = {}
  }

  on(type: string, callback: Callback, scope: Record<string, unknown> = {}) {
    if (typeof this.events[type] === 'undefined') {
      this.events[type] = []
    }
    this.events[type].push({ scope, callback })
  }

  off(type: string, callback: Callback, scope: Record<string, unknown> = {}) {
    if (typeof this.events[type] === 'undefined') {
      return
    }

    this.events[type] = this.events[type].filter((event: Event) => {
      event.scope !== scope || event.callback !== callback
    })
  }

  has(type: string, callback: Callback, scope: Record<string, unknown> = {}) {
    if (typeof this.events[type] === 'undefined') {
      return false
    }

    let numOfCallbacks = this.events[type].length
    if (callback === undefined && scope === undefined) {
      return numOfCallbacks
    }

    return this.events[type].some((event: Event) => {
      const scopeIsSame = scope ? event.scope === scope : true
      const callbackIsSame = event.callback === callback
      if (scopeIsSame && callbackIsSame) {
        return true
      }
    })
  }

  emit(type: string, ...args: any[]) {
    if (typeof this.events[type] === 'undefined') {
      return
    }

    for (const event of this.events[type]) {
      if (event && event.callback) {
        event.callback.apply(event.scope, args)
      }
    }
  }
}

export const global = new EventBus()
