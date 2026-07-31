import { objectToQueryString } from './util'

const DEFAULT_BOUNDS = {
  width: 418,
  height: 680,
  right: 50,
  bottom: 50,
}

const BOUNDS_LIMITS = {
  minWidth: 320,
  minHeight: 400,
  maxWidth: 2000,
  maxHeight: 2000,
  maxMargin: 2000,
}

const BOUNDS_STORAGE_KEY = 'zm_chat-wiki-iframe-bounds'

const RESIZE_DIRECTIONS = [
  'top',
  'right',
  'bottom',
  'left',
  'top-left',
  'top-right',
  'bottom-left',
  'bottom-right',
]

const RESIZE_CURSORS = {
  top: 'ns-resize',
  right: 'ew-resize',
  bottom: 'ns-resize',
  left: 'ew-resize',
  'top-left': 'nwse-resize',
  'top-right': 'nesw-resize',
  'bottom-left': 'nesw-resize',
  'bottom-right': 'nwse-resize',
}

function clamp(value, min, max) {
  return Math.min(max, Math.max(min, value))
}

function normalizeNumber(value, fallback, min, max) {
  const number = Number(value)
  if (!Number.isFinite(number)) {
    return fallback
  }
  return clamp(Math.round(number), min, max)
}

function isEnabled(value) {
  return value === true || value === 1 || value === '1'
}

// iframe 生命周期与跨窗口通信的唯一边界；控制器只接收已校验来源的业务消息。
class IframeBridge {
  container = null
  iframe = null
  dragHandle = null
  resizeHandles = []
  iframeSrc = ''
  boundsStorageKey = BOUNDS_STORAGE_KEY
  expectedOrigin = ''
  onMessage = null
  listening = false
  windowResizeListening = false
  activeHandle = null
  activePointerId = null
  interaction = null
  pendingPoint = null
  animationFrame = null
  previousCursor = ''
  bounds = { ...DEFAULT_BOUNDS }
  permissions = {
    drag: false,
    resize: false,
  }

  init({ iframeSrc, params, onMessage }) {
    this.iframeSrc = iframeSrc
    // 同一宿主可能接入不同机器人，按 robot_key 隔离访客调整后的窗口状态。
    this.boundsStorageKey = params?.robot_key
      ? `${BOUNDS_STORAGE_KEY}:${params.robot_key}`
      : BOUNDS_STORAGE_KEY
    this.expectedOrigin = new URL(iframeSrc).origin
    this.onMessage = onMessage

    this.insertIframe(params)

    if (!this.listening) {
      window.addEventListener('message', this.handleMessage, false)
      this.listening = true
    }

    if (!this.windowResizeListening) {
      window.addEventListener('resize', this.handleWindowResize)
      this.windowResizeListening = true
    }
  }

  insertIframe(params) {
    this.container = document.getElementById('zm_chat-wiki-iframe-container')
    const isNewContainer = !this.container

    if (!this.container) {
      this.container = document.createElement('div')
      this.container.id = 'zm_chat-wiki-iframe-container'
      this.container.style.display = 'none'
      document.body.appendChild(this.container)
    }

    this.iframe = document.getElementById('zm_chat-wiki-iframe')
    if (!this.iframe) {
      const queryStr = objectToQueryString(params)
      this.iframe = document.createElement('iframe')
      this.iframe.id = 'zm_chat-wiki-iframe'
      this.iframe.src = this.iframeSrc + '?' + queryStr
    }

    this.iframe.style.display = 'block'
    this.iframe.style.width = ''
    this.iframe.style.height = ''
    this.iframe.style.right = ''
    this.iframe.style.bottom = ''

    if (this.iframe.parentNode !== this.container) {
      this.container.prepend(this.iframe)
    }

    this.createControls()

    if (!isNewContainer) {
      this.renderBounds()
    }
  }

  createControls() {
    if (this.dragHandle || !this.container) {
      return
    }

    this.dragHandle = document.createElement('div')
    this.dragHandle.className = 'zm_chat-wiki-drag-handle'
    this.dragHandle.addEventListener('pointerdown', this.handleDragPointerDown)
    this.container.appendChild(this.dragHandle)

    this.resizeHandles = RESIZE_DIRECTIONS.map((direction) => {
      const handle = document.createElement('div')
      handle.className = `zm_chat-wiki-resize-handle zm_chat-wiki-resize-${direction}`
      handle.dataset.direction = direction
      handle.addEventListener('pointerdown', this.handleResizePointerDown)
      this.container.appendChild(handle)
      return handle
    })

    this.updateControls()
  }

  applyConfig(config = {}) {
    this.permissions.drag = isEnabled(config.iframe_drag_enabled)
    this.permissions.resize = isEnabled(config.iframe_resize_enabled)

    const configBounds = {
      width: normalizeNumber(
        config.iframe_width,
        DEFAULT_BOUNDS.width,
        BOUNDS_LIMITS.minWidth,
        BOUNDS_LIMITS.maxWidth
      ),
      height: normalizeNumber(
        config.iframe_height,
        DEFAULT_BOUNDS.height,
        BOUNDS_LIMITS.minHeight,
        BOUNDS_LIMITS.maxHeight
      ),
      right: normalizeNumber(config.iframe_right, DEFAULT_BOUNDS.right, 0, BOUNDS_LIMITS.maxMargin),
      bottom: normalizeNumber(
        config.iframe_bottom,
        DEFAULT_BOUNDS.bottom,
        0,
        BOUNDS_LIMITS.maxMargin
      ),
    }

    // 访客手动调整代表本机偏好，因此优先于管理端下发的初始边界。
    this.bounds = this.fitBounds(this.getStoredBounds() || configBounds)

    this.renderBounds()
    this.updateControls()
  }

  getStoredBounds() {
    try {
      const storedBounds = JSON.parse(window.localStorage.getItem(this.boundsStorageKey))
      const values = ['width', 'height', 'right', 'bottom'].map((key) =>
        Number(storedBounds?.[key])
      )

      if (!storedBounds || values.some((value) => !Number.isFinite(value))) {
        return null
      }

      return {
        width: normalizeNumber(
          storedBounds.width,
          DEFAULT_BOUNDS.width,
          BOUNDS_LIMITS.minWidth,
          BOUNDS_LIMITS.maxWidth
        ),
        height: normalizeNumber(
          storedBounds.height,
          DEFAULT_BOUNDS.height,
          BOUNDS_LIMITS.minHeight,
          BOUNDS_LIMITS.maxHeight
        ),
        right: normalizeNumber(
          storedBounds.right,
          DEFAULT_BOUNDS.right,
          0,
          BOUNDS_LIMITS.maxMargin
        ),
        bottom: normalizeNumber(
          storedBounds.bottom,
          DEFAULT_BOUNDS.bottom,
          0,
          BOUNDS_LIMITS.maxMargin
        ),
      }
    } catch (error) {
      return null
    }
  }

  storeBounds() {
    try {
      window.localStorage.setItem(
        this.boundsStorageKey,
        JSON.stringify({
          width: Math.round(this.bounds.width),
          height: Math.round(this.bounds.height),
          right: Math.round(this.bounds.right),
          bottom: Math.round(this.bounds.bottom),
        })
      )
    } catch (error) {
      // 宿主页面禁用本地存储时继续使用当前内存状态，不影响 SDK 交互。
    }
  }

  updateControls() {
    if (this.dragHandle) {
      this.dragHandle.style.display = this.permissions.drag ? 'block' : 'none'
    }
    this.resizeHandles.forEach((handle) => {
      handle.style.display = this.permissions.resize ? 'block' : 'none'
    })
  }

  fitBounds(bounds) {
    const viewportWidth = Math.max(1, window.innerWidth)
    const viewportHeight = Math.max(1, window.innerHeight)
    const width = Math.min(Math.max(1, bounds.width), viewportWidth)
    const height = Math.min(Math.max(1, bounds.height), viewportHeight)

    return {
      width,
      height,
      right: clamp(bounds.right, 0, Math.max(0, viewportWidth - width)),
      bottom: clamp(bounds.bottom, 0, Math.max(0, viewportHeight - height)),
    }
  }

  renderBounds() {
    if (!this.container) {
      return
    }

    this.container.style.width = `${Math.round(this.bounds.width)}px`
    this.container.style.height = `${Math.round(this.bounds.height)}px`
    this.container.style.right = `${Math.round(this.bounds.right)}px`
    this.container.style.bottom = `${Math.round(this.bounds.bottom)}px`
  }

  handleDragPointerDown = (event) => {
    if (!this.permissions.drag) {
      return
    }
    this.startInteraction(event, 'drag')
  }

  handleResizePointerDown = (event) => {
    if (!this.permissions.resize) {
      return
    }
    this.startInteraction(event, 'resize', event.currentTarget.dataset.direction)
  }

  startInteraction(event, type, direction = '') {
    if (
      !this.container ||
      event.pointerType !== 'mouse' ||
      event.button !== 0 ||
      (type === 'resize' && !RESIZE_DIRECTIONS.includes(direction))
    ) {
      return
    }

    this.endInteraction()
    event.preventDefault()

    const rect = this.container.getBoundingClientRect()
    this.interaction = {
      type,
      direction,
      startX: event.clientX,
      startY: event.clientY,
      rect: {
        left: rect.left,
        top: rect.top,
        right: rect.right,
        bottom: rect.bottom,
        width: rect.width,
        height: rect.height,
      },
    }

    this.activeHandle = event.currentTarget
    this.activePointerId = event.pointerId
    this.activeHandle.addEventListener('pointermove', this.handleInteractionMove)
    this.activeHandle.addEventListener('pointerup', this.handleInteractionEnd)
    this.activeHandle.addEventListener('pointercancel', this.handleInteractionCancel)
    this.activeHandle.addEventListener('lostpointercapture', this.handleInteractionCancel)
    this.activeHandle.setPointerCapture(this.activePointerId)

    this.previousCursor = document.documentElement.style.cursor
    document.documentElement.style.cursor =
      type === 'drag' ? 'move' : RESIZE_CURSORS[direction]
    window.addEventListener('blur', this.handleInteractionCancel)
  }

  handleInteractionMove = (event) => {
    this.pendingPoint = { clientX: event.clientX, clientY: event.clientY }
    if (this.animationFrame !== null) {
      return
    }

    this.animationFrame = window.requestAnimationFrame(() => {
      this.animationFrame = null
      if (this.pendingPoint) {
        this.applyInteraction(this.pendingPoint.clientX, this.pendingPoint.clientY)
        this.pendingPoint = null
      }
    })
  }

  handleInteractionEnd = (event) => {
    if (this.interaction) {
      this.applyInteraction(event.clientX, event.clientY)
      // 仅在正常抬起指针后保存，取消中的临时状态不覆盖上一次有效偏好。
      this.storeBounds()
    }
    this.endInteraction()
  }

  handleInteractionCancel = () => {
    this.endInteraction()
  }

  applyInteraction(clientX, clientY) {
    if (!this.interaction) {
      return
    }

    const deltaX = clientX - this.interaction.startX
    const deltaY = clientY - this.interaction.startY
    const viewportWidth = Math.max(1, window.innerWidth)
    const viewportHeight = Math.max(1, window.innerHeight)
    const start = this.interaction.rect

    let left = start.left
    let top = start.top
    let right = start.right
    let bottom = start.bottom

    if (this.interaction.type === 'drag') {
      left = clamp(start.left + deltaX, 0, Math.max(0, viewportWidth - start.width))
      top = clamp(start.top + deltaY, 0, Math.max(0, viewportHeight - start.height))
      right = left + start.width
      bottom = top + start.height
    } else {
      const direction = this.interaction.direction

      if (direction.includes('left')) {
        const minWidth = Math.min(BOUNDS_LIMITS.minWidth, start.right)
        left = clamp(
          start.left + deltaX,
          Math.max(0, start.right - BOUNDS_LIMITS.maxWidth),
          start.right - minWidth
        )
      }
      if (direction.includes('right')) {
        const availableWidth = Math.max(1, viewportWidth - start.left)
        const minWidth = Math.min(BOUNDS_LIMITS.minWidth, availableWidth)
        right = clamp(
          start.right + deltaX,
          start.left + minWidth,
          Math.min(viewportWidth, start.left + BOUNDS_LIMITS.maxWidth)
        )
      }
      if (direction.includes('top')) {
        const minHeight = Math.min(BOUNDS_LIMITS.minHeight, start.bottom)
        top = clamp(
          start.top + deltaY,
          Math.max(0, start.bottom - BOUNDS_LIMITS.maxHeight),
          start.bottom - minHeight
        )
      }
      if (direction.includes('bottom')) {
        const availableHeight = Math.max(1, viewportHeight - start.top)
        const minHeight = Math.min(BOUNDS_LIMITS.minHeight, availableHeight)
        bottom = clamp(
          start.bottom + deltaY,
          start.top + minHeight,
          Math.min(viewportHeight, start.top + BOUNDS_LIMITS.maxHeight)
        )
      }
    }

    this.bounds = {
      width: Math.max(1, right - left),
      height: Math.max(1, bottom - top),
      right: Math.max(0, viewportWidth - right),
      bottom: Math.max(0, viewportHeight - bottom),
    }
    this.renderBounds()
  }

  endInteraction() {
    if (this.animationFrame !== null) {
      window.cancelAnimationFrame(this.animationFrame)
      this.animationFrame = null
    }

    this.pendingPoint = null
    this.interaction = null
    window.removeEventListener('blur', this.handleInteractionCancel)

    if (this.activeHandle) {
      document.documentElement.style.cursor = this.previousCursor
      this.previousCursor = ''
      this.activeHandle.removeEventListener('pointermove', this.handleInteractionMove)
      this.activeHandle.removeEventListener('pointerup', this.handleInteractionEnd)
      this.activeHandle.removeEventListener('pointercancel', this.handleInteractionCancel)
      this.activeHandle.removeEventListener('lostpointercapture', this.handleInteractionCancel)

      if (
        this.activePointerId !== null &&
        this.activeHandle.hasPointerCapture(this.activePointerId)
      ) {
        this.activeHandle.releasePointerCapture(this.activePointerId)
      }
    }

    this.activeHandle = null
    this.activePointerId = null
  }

  handleWindowResize = () => {
    this.endInteraction()
    this.bounds = this.fitBounds(this.bounds)
    this.renderBounds()
  }

  handleMessage = (event) => {
    if (!event.origin) {
      return
    }

    let eventOrigin = ''
    try {
      eventOrigin = new URL(event.origin).origin
    } catch (error) {
      return
    }

    // 只把目标 iframe 来源的消息交给控制器，避免宿主页面其他消息干扰 SDK 状态。
    if (eventOrigin !== this.expectedOrigin) {
      return
    }

    if (typeof this.onMessage === 'function') {
      this.onMessage(event.data)
    }
  }

  show() {
    if (!this.container) {
      return false
    }

    this.container.style.display = 'block'
    return true
  }

  hide() {
    if (this.container) {
      this.container.style.display = 'none'
    }
  }

  send(action, data) {
    if (data) {
      try {
        data = JSON.parse(JSON.stringify(data))
      } catch (error) {
        console.error('Failed to stringify data:', error)
        return
      }
    }

    if (this.iframe?.contentWindow && typeof this.iframe.contentWindow.postMessage === 'function') {
      try {
        // 保持现有跨域通信行为；targetOrigin 安全加固需要 SDK 与 iframe 两端同步调整。
        this.iframe.contentWindow.postMessage({ action, data }, '*')
      } catch (error) {
        console.error('Failed to post message:', error)
      }
    } else {
      console.warn('frame.contentWindow is not available or postMessage is not supported.')
    }
  }

  remove() {
    this.endInteraction()

    if (this.dragHandle) {
      this.dragHandle.removeEventListener('pointerdown', this.handleDragPointerDown)
    }
    this.resizeHandles.forEach((handle) => {
      handle.removeEventListener('pointerdown', this.handleResizePointerDown)
    })

    if (this.listening) {
      window.removeEventListener('message', this.handleMessage, false)
      this.listening = false
    }

    if (this.windowResizeListening) {
      window.removeEventListener('resize', this.handleWindowResize)
      this.windowResizeListening = false
    }

    if (this.container?.parentNode) {
      this.container.parentNode.removeChild(this.container)
    } else if (this.iframe?.parentNode) {
      this.iframe.parentNode.removeChild(this.iframe)
    }

    // 清空跨实例引用，避免未来重新初始化时继续调用旧控制器。
    this.container = null
    this.iframe = null
    this.dragHandle = null
    this.resizeHandles = []
    this.iframeSrc = ''
    this.boundsStorageKey = BOUNDS_STORAGE_KEY
    this.expectedOrigin = ''
    this.onMessage = null
    this.bounds = { ...DEFAULT_BOUNDS }
    this.permissions = { drag: false, resize: false }
  }
}

export default new IframeBridge()
