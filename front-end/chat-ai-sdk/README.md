# Chat AI SDK

用于在第三方网页中嵌入 ChatWiki 会话窗口。SDK 会创建一个隐藏的 iframe，并提供打开、关闭、运行时配置和生命周期订阅能力。

## 快速接入

将下面的代码放到接入页面中，并替换 `robot_key` 与 SDK 地址：

```html
<script>
  ;(function (window, document) {
    const script = document.createElement('script')
    const firstScript = document.getElementsByTagName('script')[0]

    script.async = true
    script.charset = 'UTF-8'
    script.id = 'ai_chat_js'
    script.setAttribute(
      'data-json',
      JSON.stringify({
        robot_key: 'YOUR_ROBOT_KEY',
        language: 'zh-CN',
        show_float_button: true
      })
    )
    script.src = 'https://YOUR_DOMAIN/sdk/ai-chat-sdk.umd.cjs'
    firstScript.parentNode.insertBefore(script, firstScript)
  })(window, document)
</script>
```

SDK 文件执行完成后会创建：

```js
window.AiChatSDK
```

`AiChatSDK` 创建后即可调用 `open()`；如果 iframe 尚未初始化完成，SDK 会保存打开请求，并在 ready 后自动执行。

## 初始化配置

初始化配置通过 SDK script 的 `data-json` 传入。

| 字段 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `robot_key` | `string` | - | 机器人标识 |
| `openid` | `string` | 自动生成 | 外部用户唯一标识 |
| `nickname` | `string` | `''` | 外部用户昵称 |
| `name` | `string` | `''` | 外部用户姓名 |
| `avatar` | `string` | `''` | 外部用户头像 URL |
| `language` | `string` | `zh-CN` | 会话语言 |
| `show_float_button` | `boolean` | `true` | 是否显示 SDK 默认悬浮按钮 |

`show_float_button` 是 SDK 专属配置，不会传入 iframe 查询参数。

### iframe 参数传递逻辑

SDK 初始化时读取 `id="ai_chat_js"` 脚本标签上的 `data-json`，并按以下链路传递参数：

```text
script data-json
  → SDK 解析初始化配置
  → 移除 show_float_button 等 SDK 专属字段
  → 其余字段序列化为 iframe query
  → chat-ai-pc 从 route.query 读取
  → 创建或恢复会话
```

例如：

```js
script.setAttribute(
  'data-json',
  JSON.stringify({
    robot_key: 'YOUR_ROBOT_KEY',
    openid: 'user_10001',
    nickname: '小明',
    name: '张三',
    avatar: 'https://example.com/avatar.png',
    language: 'zh-CN',
    show_float_button: true
  })
)
```

其中 `openid`、`nickname`、`name`、`avatar` 会经过 URL 编码后传入 iframe；`show_float_button` 只由 SDK 消费。未传 `openid` 时，`chat-ai-pc` 会使用本地生成的匿名标识。

这套机制只用于 iframe 初始化。`AiChatSDK.open(options)` 的参数通过 `postMessage` 发送，用于打开窗口时的业务选项，不会替代初始化用户信息。

### 添加新的初始化参数

添加普通 iframe 初始化参数时：

1. 接入方在 `data-json` 中增加字段，SDK 的通用序列化逻辑通常无需修改。
2. 在 `chat-ai-pc` 的 `ChatInitParams` 中声明字段。
3. 在 `parseChatInitParams()` 中统一读取 query、设置默认值并完成类型转换。
4. 如果字段需要进入用户状态或后端接口，再同步修改 Store 赋值和对应接口请求参数。
5. 如果字段只供 SDK 自身使用，应在 SDK 初始化阶段消费并从 iframe 参数中移除，参考 `show_float_button`。

初始化参数会出现在 iframe URL 中，不建议通过该机制传递令牌、手机号等敏感信息。

## 自定义入口

隐藏默认悬浮按钮，并使用接入页面自己的按钮打开会话：

```html
<button id="open-ai-chat">在线咨询</button>

<script>
  document.getElementById('open-ai-chat').addEventListener('click', function () {
    if (!window.AiChatSDK) {
      return
    }

    window.AiChatSDK.setConfig('showFloatButton', false)
    window.AiChatSDK.open({ source: 'custom-button' })
  })
</script>
```

调用 SDK 方法前，`window.AiChatSDK` 必须已经存在；SDK 不处理脚本文件尚未执行时的调用。

## API

### `open(options?)`

打开会话窗口。`options` 必须是可序列化的普通对象，会通过 `openWindow` 消息透传给 iframe。

```js
AiChatSDK.open()

AiChatSDK.open({
  source: 'pricing-page'
})
```

iframe 未 ready 时调用会进入等待状态；等待期间多次调用以最后一次参数为准。窗口已经打开时重复调用不会重复触发。

### `close()`

关闭会话窗口，并根据 `showFloatButton` 决定是否恢复悬浮按钮。

```js
AiChatSDK.close()
```

### `setConfig(key, value)` / `setConfig(config)`

更新 SDK 运行时配置。运行时配置优先级高于 `data-json`。

```js
AiChatSDK.setConfig('showFloatButton', false)
```

```js
AiChatSDK.setConfig({
  showFloatButton: true
})
```

当前支持的运行时配置：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `showFloatButton` | `boolean` | 动态显示或隐藏默认悬浮按钮 |

隐藏悬浮按钮时，SDK 会同时清除未读红点和新消息提示。

### `on(event, callback)`

订阅 SDK 生命周期事件，返回取消监听函数。

```js
const offOpen = AiChatSDK.on('open', function (event) {
  console.log('会话已打开', event.options)
})

const offClose = AiChatSDK.on('close', function (event) {
  console.log('会话已关闭', event.source)
})

offOpen()
offClose()
```

支持的事件：

| 事件 | 回调参数 | 触发时机 |
| --- | --- | --- |
| `ready` | `{ type: 'ready' }` | iframe 完成业务初始化 |
| `open` | `{ type: 'open', options }` | 会话窗口实际打开 |
| `close` | `{ type: 'close', source }` | 会话窗口实际关闭 |

`close.source` 为 `sdk` 或 `iframe`，分别表示通过 SDK 主动关闭或在会话窗口内关闭。

`ready` 是可回放的一次性状态事件；ready 后订阅仍会异步执行一次。`open` 和 `close` 不回放历史状态。

### `onReady(callback)`

`on('ready', callback)` 的便捷写法，同样返回取消监听函数。

```js
const offReady = AiChatSDK.onReady(function () {
  console.log('Chat AI SDK ready')
})

offReady()
```

### `isReady()`

同步返回 iframe 是否已经完成初始化。

```js
if (AiChatSDK.isReady()) {
  console.log('ready')
}
```

不建议通过轮询 `isReady()` 等待初始化；需要等待时使用 `onReady()`。

## 完整示例

```js
const sdk = window.AiChatSDK

const offReady = sdk.onReady(function () {
  console.log('ready')
})

const offOpen = sdk.on('open', function (event) {
  console.log('open', event.options)
})

const offClose = sdk.on('close', function (event) {
  console.log('close', event.source)
})

sdk.setConfig({ showFloatButton: false })
sdk.open({ source: 'page-action' })

window.addEventListener('beforeunload', function () {
  offReady()
  offOpen()
  offClose()
}, { once: true })
```

## iframe 大小与位置控制

嵌入窗口的尺寸、位置、拖拽和缩放能力由 SDK 宿主层统一控制。`chat-ai-pc` 只负责 iframe 内的会话页面和业务消息，不负责监听或转发拖拽过程中的鼠标事件。

### 配置来源

以下字段配置在管理端的 `external_config_pc` 中，由 `chat-ai-pc` 初始化后通过现有的 `init` 消息传给 SDK：

| 字段 | 默认值 | 允许范围 | 说明 |
| --- | --- | --- | --- |
| `iframe_width` | `418` | `320`～`2000` | iframe 容器宽度，单位为 px |
| `iframe_height` | `680` | `400`～`2000` | iframe 容器高度，单位为 px |
| `iframe_right` | `50` | `0`～`2000` | 容器距离浏览器视口右侧的距离，单位为 px |
| `iframe_bottom` | `50` | `0`～`2000` | 容器距离浏览器视口底部的距离，单位为 px |
| `iframe_drag_enabled` | `false` | 布尔值 | 是否允许拖动窗口 |
| `iframe_resize_enabled` | `false` | 布尔值 | 是否允许从四边和四角调整窗口大小 |

这些字段属于机器人外部服务配置，不是 SDK `data-json` 初始化参数，也不能通过当前的 `AiChatSDK.setConfig()` 动态修改。SDK 收到配置后会进行数值转换、范围限制和视口适配；旧数据缺少字段或字段无效时使用默认值。

### 宿主层 DOM 结构

SDK 创建的结构如下：

```text
#zm_chat-wiki-iframe-container
├── #zm_chat-wiki-iframe
├── .zm_chat-wiki-drag-handle
└── .zm_chat-wiki-resize-handle × 8
```

外层容器使用 `position: fixed`，实际的 `width`、`height`、`right`、`bottom` 都设置在该容器上；iframe 使用 `width: 100%` 和 `height: 100%` 跟随容器。拖拽区域和八个缩放手柄是宿主页面中的透明覆盖层，与 iframe 同级，因此交互事件从一开始就发生在宿主页面。

### Pointer Events 交互流程

1. 用户在拖拽区域或缩放手柄按下鼠标，宿主层监听 `pointerdown`，记录指针起点和容器初始矩形。
2. 当前手柄调用 `setPointerCapture(pointerId)` 捕获该指针。即使指针随后移出手柄或跨过 iframe，后续 `pointermove` 和 `pointerup` 仍会发送给该宿主元素。
3. `pointermove` 使用 `clientX`、`clientY` 计算相对位移，并通过 `requestAnimationFrame` 合并高频更新。
4. 拖拽保持宽高不变并限制窗口完全位于视口内；缩放根据手柄方向调整对应边，应用最小尺寸、最大尺寸和视口边界限制。
5. `pointerup`、`pointercancel`、`lostpointercapture` 或浏览器窗口失焦时结束交互，释放 Pointer Capture、事件监听和临时光标状态。

当前只响应鼠标左键，即 `pointerType === 'mouse'` 且 `button === 0`；触摸和触控笔尚未纳入支持范围。

### 本地缓存与优先级

用户通过拖拽或缩放完成一次有效调整后，SDK 会将容器的 `width`、`height`、`right`、`bottom` 写入宿主页面域名下的 `localStorage`。缓存键为 `zm_chat-wiki-iframe-bounds:<robot_key>`；缺少 `robot_key` 时使用 `zm_chat-wiki-iframe-bounds`，避免不同机器人共用同一份窗口状态。

窗口尺寸和位置的取值优先级为：本地有效缓存 > 管理端 `external_config_pc` > SDK 默认值。本地缓存只覆盖窗口边界，`iframe_drag_enabled` 和 `iframe_resize_enabled` 仍始终使用管理端最新配置。缓存内容会按 SDK 的尺寸、距离和当前视口限制再次修正；缓存缺失、字段无效、JSON 损坏或浏览器禁用本地存储时，SDK 会安全回退到管理端配置。

只有正常触发 `pointerup` 时才写入缓存；`pointercancel`、`lostpointercapture` 或窗口失焦造成的中断不会用临时状态覆盖上一次有效缓存。SDK 销毁或关闭窗口不会删除缓存。

### 状态与边界处理

- 同一页面中关闭后再次打开窗口，会保留本次运行期间拖拽或缩放后的内存状态。
- 刷新或重新打开宿主页面后，会优先恢复当前机器人对应的本地尺寸与位置；没有有效缓存时才使用管理端配置。
- 浏览器视口变化时，SDK 会结束正在进行的交互，并重新限制窗口尺寸和位置，避免窗口落到可视区域外。
- 管理端配置大于当前视口时，实际显示尺寸会缩小到视口范围；右侧和底部距离也会按可用空间修正。

### 实现注意事项

- 不需要把 iframe 内的 `mousedown`、`mousemove`、`mouseup` 或对应 Pointer Events 通过 `postMessage` 转发给宿主页面。iframe 事件不会冒泡到宿主文档，跨窗口转发还会增加事件丢失、节流、坐标换算和结束状态清理的复杂度。
- Pointer Capture 必须由接收到 `pointerdown` 的宿主层手柄发起。不要只在 iframe 内按下后，尝试由宿主页面补捕获该指针。
- 顶部拖拽区域当前为 `top: 6px; left: 6px; right: 40px; height: 58px`，右侧预留了 iframe 内关闭按钮的点击区域。如果会话页头部新增按钮或调整布局，需要同步检查并修改宿主层拖拽区域，避免透明覆盖层拦截点击。
- 八个缩放手柄覆盖容器四边和四角。修改手柄厚度、圆角或容器结构时，应同时检查可点击区域、光标样式和几何计算是否一致。
- `IframeBridge` 是宿主层 iframe 容器、交互监听和边界状态的唯一管理入口。新增相关能力时应继续在该边界内处理，不要在 `chat-ai-pc` 重复维护外层位置状态。
- 新增交互结束路径时，要与现有 `pointerup`、`pointercancel`、`lostpointercapture`、`blur`、窗口尺寸变化和 SDK 销毁逻辑保持一致，确保监听器、动画帧和 Pointer Capture 都能清理。

## 本地 Demo

本地实时调试需要同时启动 `chat-ai-pc`。SDK 的开发环境会通过 iframe 加载 `chat-ai-pc`，当前 `.env.development` 默认使用 `http://localhost:5173`：

```bash
# 终端 1：启动 iframe 会话页面
cd ../chat-ai-pc
npm install
npm run dev
```

确认 `chat-ai-pc` 已运行后，再启动 SDK Demo：

```bash
# 终端 2：启动 SDK Demo
cd ../chat-ai-sdk
npm install
npm run dev
```

SDK Demo 固定使用 `http://localhost:5180`，并启用了 `strictPort`。如果 5180 已被占用，Vite 会直接退出，不会自动切换到其他端口；需要先释放该端口再重新启动。

开始调试前，还需要将根目录 `index.html` 中 `data-json` 配置的 `robot_key` 替换为当前环境中实际可用的机器人标识。

如果 `chat-ai-pc` 使用了其他端口或地址，需要同步修改 SDK `.env.development` 中的 `VITE_AI_CHAT_BASE_URL`。未启动 `chat-ai-pc` 或地址不一致时，iframe 无法完成初始化，`onReady()` 不会触发，ready 前调用的 `open()` 也会继续等待。

Demo 包含：

- 自定义按钮打开和关闭会话。
- 动态显示、隐藏悬浮按钮。
- ready、open、close 状态与事件日志。
- `open(options)` 参数透传示例。

`public/index.html` 提供了一份不依赖 Vue 的静态接入示例。

### SDK 接入沙箱

`public/sdk-test.html` 用于粘贴并运行完整的 SDK 接入代码。测试工具的面板和预览布局位于外层页面，接入代码运行在带有 sandbox 权限的预览 iframe 中，避免 SDK 的固定定位元素和宿主 DOM 直接影响测试工具布局。

预览 iframe 使用同源 `srcdoc` 装载模拟宿主页，并保留 `allow-scripts`、`allow-same-origin` 等现有权限。这样既保持原有预览区域、消息通道和控制按钮不变，也使 iframe 内的 SDK 能够像真实接入页面一样访问当前测试站点的 `localStorage`。不要改回 `data:` URL，否则预览文档会使用独立来源，浏览器可能拒绝其中的本地存储访问。

测试流程如下：

1. 在编辑区粘贴接入代码并点击“运行测试”，代码会保存到测试站点的 `localStorage`。
2. SDK 在同源预览 iframe 中初始化；正常结束拖拽或缩放后，窗口边界按 `robot_key` 写入本地缓存。
3. 刷新测试页面后，编辑区会恢复上次接入代码，现有 `load` 监听会自动重新运行。
4. 新预览 iframe 与刷新前保持同源，SDK 会优先恢复对应机器人的本地窗口边界。

本地缓存仍遵循浏览器的来源隔离规则，协议、域名或端口任一变化都会进入另一份存储空间。例如 `http://localhost:5180` 与其他端口之间不会共享接入代码或窗口边界缓存。点击“清空预览”只销毁当前 iframe，不删除缓存；“恢复默认”只重置接入代码，不清理 SDK 的窗口边界缓存。
