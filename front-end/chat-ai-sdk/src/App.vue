<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const sdkLoaded = ref(false)
const sdkReady = ref(false)
const showFloatButton = ref(false)
const logs = ref([])
const cleanupListeners = []
let logId = 0
let sdkScript = null
let sdkRegistered = false

const statusText = computed(() => {
  if (!sdkLoaded.value) {
    return '等待 SDK 文件'
  }
  return sdkReady.value ? '会话已就绪' : '等待 iframe 初始化'
})

const addLog = (type, detail) => {
  logs.value.unshift({
    id: ++logId,
    type,
    detail,
    time: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
  })
  logs.value = logs.value.slice(0, 8)
}

const getSdk = () => window.AiChatSDK

const registerSdk = () => {
  const sdk = getSdk()
  if (!sdk || sdkRegistered) {
    return
  }

  sdkRegistered = true
  sdkLoaded.value = true
  sdkReady.value = sdk.isReady()
  addLog('SDK', '公开 API 已挂载到 window.AiChatSDK')

  cleanupListeners.push(
    sdk.onReady(() => {
      sdkReady.value = true
      addLog('READY', 'iframe 业务初始化完成')
    }),
    sdk.on('open', (event) => {
      addLog('OPEN', JSON.stringify(event.options))
    }),
    sdk.on('close', (event) => {
      addLog('CLOSE', `source: ${event.source}`)
    }),
  )
}

const openChat = () => {
  const sdk = getSdk()
  if (!sdk) {
    addLog('WARN', 'SDK 文件尚未加载')
    return
  }

  sdk.open({
    source: 'vite-demo',
    trigger: 'primary-action',
  })
}

const closeChat = () => {
  const sdk = getSdk()
  if (!sdk) {
    addLog('WARN', 'SDK 文件尚未加载')
    return
  }
  sdk.close()
}

const updateFloatButton = () => {
  const sdk = getSdk()
  if (!sdk) {
    showFloatButton.value = !showFloatButton.value
    addLog('WARN', 'SDK 文件尚未加载')
    return
  }

  sdk.setConfig('showFloatButton', showFloatButton.value)
  addLog('CONFIG', `showFloatButton: ${showFloatButton.value}`)
}

onMounted(() => {
  registerSdk()

  if (!sdkRegistered) {
    sdkScript = document.getElementById('ai_chat_js')
    sdkScript?.addEventListener('load', registerSdk, { once: true })
    addLog('BOOT', '等待异步 SDK 脚本执行')
  }
})

onBeforeUnmount(() => {
  sdkScript?.removeEventListener('load', registerSdk)
  cleanupListeners.forEach((cleanup) => cleanup())
})
</script>

<template>
  <main class="demo-shell">
    <header class="topbar">
      <div class="brand-mark">CA</div>
      <div class="brand-name">Chat AI SDK Demo</div>
      <div class="status-pill" :class="{ ready: sdkReady }">
        <span class="status-dot"></span>
        {{ statusText }}
      </div>
    </header>

    <section class="demo-panel">
      <div class="control-panel">
        <h1>会话控制</h1>
        <div class="actions">
          <button class="primary-action" :disabled="!sdkLoaded" @click="openChat">
            打开会话
            <span>↗</span>
          </button>
          <button class="secondary-action" :disabled="!sdkLoaded" @click="closeChat">
            关闭会话
          </button>
        </div>

        <label class="config-toggle">
          <strong>显示默认悬浮按钮</strong>
          <input v-model="showFloatButton" type="checkbox" :disabled="!sdkLoaded" @change="updateFloatButton" />
          <span class="toggle-track"></span>
        </label>
      </div>

      <aside class="event-panel">
        <div class="panel-heading">
          <h2>事件日志</h2>
          <span class="event-count">{{ logs.length.toString().padStart(2, '0') }}</span>
        </div>

        <div class="event-list" aria-live="polite">
          <div v-for="log in logs" :key="log.id" class="event-row">
            <time>{{ log.time }}</time>
            <span class="event-type">{{ log.type }}</span>
            <span class="event-detail">{{ log.detail }}</span>
          </div>
          <div v-if="logs.length === 0" class="empty-log">等待 SDK 事件</div>
        </div>
      </aside>
    </section>
  </main>
</template>

<style scoped>
.demo-shell {
  width: min(920px, calc(100% - 32px));
  margin: 0 auto;
  padding: 24px 0;
}

.topbar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--ink);
}

.brand-mark {
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  color: var(--paper);
  background: var(--ink);
  font-family: var(--mono);
  font-weight: 700;
}

.brand-name {
  font-family: var(--display);
  font-size: 16px;
  font-weight: 700;
}

.status-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
  padding: 6px 10px;
  border: 1px solid var(--ink);
  font-family: var(--mono);
  font-size: 12px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--orange);
  box-shadow: 0 0 0 4px rgba(255, 91, 41, 0.13);
}

.status-pill.ready .status-dot {
  background: var(--green);
  box-shadow: 0 0 0 4px rgba(29, 151, 108, 0.13);
}

.demo-panel {
  display: grid;
  grid-template-columns: minmax(280px, 0.8fr) minmax(360px, 1.2fr);
  border-bottom: 1px solid var(--ink);
}

.control-panel {
  padding: 28px 28px 28px 0;
  border-right: 1px solid var(--ink);
}

h1 {
  margin: 0;
  font-family: var(--display);
  font-size: 24px;
  font-weight: 700;
}

.actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

button {
  min-height: 40px;
  padding: 0 16px;
  border: 1px solid var(--ink);
  border-radius: 0;
  font: 600 14px var(--body);
  cursor: pointer;
  transition: transform 160ms ease, box-shadow 160ms ease, background 160ms ease;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.4;
}

button:not(:disabled):hover {
  transform: translate(-2px, -2px);
  box-shadow: 4px 4px 0 var(--ink);
}

.primary-action {
  display: flex;
  align-items: center;
  gap: 24px;
  color: #fff;
  background: var(--blue);
}

.primary-action span {
  font-size: 20px;
}

.secondary-action {
  color: var(--ink);
  background: transparent;
}

.config-toggle {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid rgba(22, 24, 29, 0.2);
  cursor: pointer;
}

.config-toggle strong {
  font-size: 14px;
}

.config-toggle input {
  position: absolute;
  opacity: 0;
}

.toggle-track {
  position: relative;
  flex: 0 0 auto;
  width: 48px;
  height: 26px;
  border: 1px solid var(--ink);
  background: #fff;
  transition: background 160ms ease;
}

.toggle-track::after {
  content: '';
  position: absolute;
  top: 4px;
  left: 4px;
  width: 16px;
  height: 16px;
  background: var(--ink);
  transition: transform 160ms ease, background 160ms ease;
}

.config-toggle input:checked + .toggle-track {
  background: var(--blue);
}

.config-toggle input:checked + .toggle-track::after {
  background: #fff;
  transform: translateX(22px);
}

.event-panel {
  min-width: 0;
  padding: 28px 0 28px 28px;
}

.panel-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

h2 {
  margin: 0;
  font-family: var(--display);
  font-size: 24px;
}

.event-count {
  font: 700 22px var(--mono);
  color: var(--blue);
}

.event-list {
  border-top: 1px solid var(--ink);
}

.event-row {
  display: grid;
  grid-template-columns: 70px 58px minmax(0, 1fr);
  gap: 12px;
  padding: 11px 0;
  border-bottom: 1px solid rgba(22, 24, 29, 0.16);
  font-family: var(--mono);
  font-size: 11px;
  animation: reveal 240ms ease both;
}

.event-row time {
  color: var(--muted);
}

.event-type {
  font-weight: 700;
  color: var(--blue);
}

.event-detail {
  overflow: hidden;
  color: var(--ink);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-log {
  padding: 24px 0;
  font-family: var(--mono);
  font-size: 12px;
  color: var(--muted);
}

@keyframes reveal {
  from {
    opacity: 0;
    transform: translateY(-4px);
  }
}

@media (max-width: 820px) {
  .demo-shell {
    width: min(100% - 28px, 640px);
  }

  .status-pill {
    max-width: 150px;
  }

  .demo-panel {
    grid-template-columns: 1fr;
  }

  .control-panel {
    padding: 24px 0;
    border-right: 0;
  }

  .event-panel {
    padding: 24px 0;
    border-top: 1px solid var(--ink);
  }
}

@media (max-width: 520px) {
  .topbar {
    align-items: flex-start;
  }

  .status-pill {
    padding: 6px 8px;
    font-size: 10px;
  }

  .actions {
    flex-direction: column;
  }

  .primary-action,
  .secondary-action {
    justify-content: space-between;
    width: 100%;
  }

  .event-row {
    grid-template-columns: 62px 52px minmax(0, 1fr);
    gap: 8px;
  }
}
</style>
