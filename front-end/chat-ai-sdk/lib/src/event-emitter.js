export default class EventEmitter {
  supportedEvents = new Set()
  listeners = new Map()

  constructor(supportedEvents = []) {
    supportedEvents.forEach((event) => {
      this.supportedEvents.add(event)
      this.listeners.set(event, new Set())
    })
  }

  on(event, callback, options = {}) {
    if (!this.supportedEvents.has(event)) {
      console.warn(`Unsupported event: ${event}`)
      return () => {}
    }

    if (typeof callback !== 'function') {
      console.warn('Event callback must be a function.')
      return () => {}
    }

    // replayPayload 只用于已经发生的状态事件异步回放，不会保存为常驻监听器。
    if (Object.prototype.hasOwnProperty.call(options, 'replayPayload')) {
      let active = true
      queueMicrotask(() => {
        if (active) {
          this.callListener(event, callback, options.replayPayload)
        }
      })

      return () => {
        active = false
      }
    }

    const eventListeners = this.listeners.get(event)
    eventListeners.add(callback)

    return () => {
      eventListeners.delete(callback)
    }
  }

  emit(event, payload) {
    const eventListeners = this.listeners.get(event)

    if (!eventListeners) {
      console.warn(`Unsupported event: ${event}`)
      return
    }

    // 使用快照，避免监听器在派发过程中增删订阅影响当前轮次。
    Array.from(eventListeners).forEach((callback) => {
      this.callListener(event, callback, payload)
    })
  }

  clear(event) {
    if (event === undefined) {
      this.listeners.forEach((eventListeners) => eventListeners.clear())
      return
    }

    const eventListeners = this.listeners.get(event)
    if (!eventListeners) {
      console.warn(`Unsupported event: ${event}`)
      return
    }

    eventListeners.clear()
  }

  callListener(event, callback, payload) {
    try {
      callback(payload)
    } catch (error) {
      console.error(`Failed to handle SDK event: ${event}`, error)
    }
  }
}
