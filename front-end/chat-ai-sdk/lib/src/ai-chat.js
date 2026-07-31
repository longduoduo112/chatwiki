import AiAvatar from './ai-avatar'
import AiDot from './ai-dot'
import NewMessage from './new-message'
import IframeBridge from './iframe-bridge'
import EventEmitter from './event-emitter'

const supportedEvents = ['ready', 'open', 'close']

function isPlainObject(value) {
  return Object.prototype.toString.call(value) === '[object Object]'
}

class AiChatWidget {
  config = {}
  ready = false
  opened = false
  desiredOpen = false
  pendingOpenOptions = {}
  initData = null
  runtimeConfig = {
    showFloatButton: true,
  }
  eventEmitter = new EventEmitter(supportedEvents)

  constructor() {
    
  }

  init(config) {
    this.config = config;
    this.runtimeConfig.showFloatButton = this.config.showFloatButton !== false

    IframeBridge.init({
      iframeSrc: this.config.iframeSrc,
      params: this.config.params,
      onMessage: this.handleMessage.bind(this),
    })

    return {
      open: this.open.bind(this),
      close: this.close.bind(this),
      setConfig: this.setConfig.bind(this),
      on: this.on.bind(this),
      onReady: this.onReady.bind(this),
      isReady: this.isReady.bind(this),
    };
  }

  removeAiChat() {
    IframeBridge.remove()
  }

  onInit(data) {
    this.initData = data
    IframeBridge.applyConfig(data?.config)
    AiAvatar.init(data, this.runtimeConfig.showFloatButton, () => this.open())

    if (this.desiredOpen) {
      AiAvatar.hide()
    }

    if (this.ready) {
      return
    }

    this.ready = true
    // 先通知接入方 ready，再执行待打开请求，保证 ready → open 的事件顺序稳定。
    this.eventEmitter.emit('ready', { type: 'ready' })
    // ready 状态由控制器保留，首次派发后清理一次性监听器，后续订阅通过状态回放。
    this.eventEmitter.clear('ready')

    if (this.desiredOpen) {
      this.performOpen(this.pendingOpenOptions)
    }
  }

  open(options = {}) {
    options = this.normalizeOpenOptions(options)

    if (this.opened) {
      return
    }

    if (!this.ready) {
      this.desiredOpen = true
      this.pendingOpenOptions = options
      return
    }

    this.performOpen(options)
  }

  performOpen(options) {
    if (this.opened || !IframeBridge.show()) {
      return
    }

    this.desiredOpen = false
    this.pendingOpenOptions = {}
    AiAvatar.hide()
    IframeBridge.send('openWindow', options)
    this.opened = true
    this.eventEmitter.emit('open', { type: 'open', options })
  }

  close() {
    this.desiredOpen = false
    this.pendingOpenOptions = {}
    IframeBridge.hide()

    if (!this.opened) {
      AiAvatar.show()
      return
    }

    this.opened = false
    IframeBridge.send('closeWindow', {})
    AiAvatar.show();
    this.eventEmitter.emit('close', { type: 'close', source: 'sdk' })
  }

  onClose() {
    const wasOpened = this.opened
    this.desiredOpen = false
    this.pendingOpenOptions = {}
    this.opened = false
    AiAvatar.show()
    IframeBridge.hide()

    if (wasOpened) {
      this.eventEmitter.emit('close', { type: 'close', source: 'iframe' })
    }
  }

  setConfig(keyOrConfig, value) {
    let nextConfig = keyOrConfig

    if (typeof keyOrConfig === 'string') {
      if (arguments.length < 2) {
        console.warn('setConfig requires a value when called with a key.')
        return
      }
      nextConfig = { [keyOrConfig]: value }
    }

    if (!isPlainObject(nextConfig)) {
      console.warn('setConfig requires a key/value pair or a plain object.')
      return
    }

    Object.keys(nextConfig).forEach((key) => {
      if (key !== 'showFloatButton') {
        console.warn(`Unsupported config: ${key}`)
        return
      }

      if (typeof nextConfig[key] !== 'boolean') {
        console.warn('showFloatButton must be a boolean.')
        return
      }

      this.runtimeConfig.showFloatButton = nextConfig[key]
      AiAvatar.setEnabled(nextConfig[key])

      if (!nextConfig[key]) {
        AiDot.remove()
        NewMessage.remove()
      } else if (this.ready && !this.opened) {
        AiAvatar.show()
      }
    })
  }

  on(event, callback) {
    if (event === 'ready' && this.ready) {
      return this.eventEmitter.on(event, callback, {
        replayPayload: { type: 'ready' },
      })
    }

    return this.eventEmitter.on(event, callback)
  }

  onReady(callback) {
    return this.on('ready', callback)
  }

  isReady() {
    return this.ready
  }

  normalizeOpenOptions(options) {
    if (!isPlainObject(options)) {
      console.warn('open options must be a plain object.')
      return {}
    }

    try {
      return JSON.parse(JSON.stringify(options))
    } catch (error) {
      console.warn('open options must be serializable.', error)
      return {}
    }
  }

  createDot(data) {
    if (!this.runtimeConfig.showFloatButton) {
      AiDot.remove()
      return
    }

    AiDot.create({value: data})
  }

  createNewMessage(data) {
    if (!this.runtimeConfig.showFloatButton) {
      NewMessage.remove()
      return
    }

    let list = data || [];
    
    if(list.length === 0){
      NewMessage.remove()
      return
    }

    // 截取数组的最后一个元素
    list = list.slice(-1);

    NewMessage.create(list)
  }

  handleMessage(res) {
    if(import.meta.env.DEV){
      console.log('Received message from iframe:', res);
    }
    
    if(!res){
      return
    }

    if (res.action === "closeChat") {
      this.onClose();
    }

    if (res.action === "init") {
      this.onInit(res.data)
    }
    // 更新未读消息数
    if (res.action === "dot") {
      this.createDot(res.data);
    }
    // 更新未读消息
    if (res.action === "newMessage") {
      this.createNewMessage(res.data);
    }
  }
}

export default new AiChatWidget();
