# chat-ai-mobile

Chat AI 会话页面，可独立访问，也可作为 `chat-ai-sdk` 创建的 iframe 页面运行。

## SDK iframe 参数传递

### 传递链路

接入页面通过 SDK 脚本的 `data-json` 提供初始化参数。SDK 将非 SDK 专属字段序列化为 iframe query，当前项目通过统一入口解析：

```text
chat-ai-sdk data-json
  → iframe URL query
  → route.query
  → parseChatInitParams()
  → ChatInitParams
  → Chat Store
  → POST /chat/welcome
```

相关代码：

- `src/types/chat.ts`：定义允许进入会话初始化流程的参数类型。
- `src/utils/chat-init.ts`：统一读取 query、设置默认值并转换类型。
- `src/router/router-guards.ts`：路由阶段读取相同参数并预取机器人展示及语言配置。
- `src/views/chat/index.vue`：聊天页面读取相同参数，确定历史会话后完成初始化。
- `src/stores/modules/chat.ts`：保存用户和机器人状态，并组装 `/chat/welcome` 请求参数。

当前初始化参数：

| 字段 | 类型 | 默认值 | 用途 |
| --- | --- | --- | --- |
| `robot_key` | `string` | `''` | 机器人标识 |
| `openid` | `string` | 本地匿名标识 | 外部用户唯一标识 |
| `nickname` | `string` | `''` | 用户昵称 |
| `name` | `string` | `''` | 用户姓名 |
| `avatar` | `string` | `''` | 用户头像 URL |
| `dialogue_id` | `number` | `0` | 指定需要恢复的会话 |

`isOpen`、`unreadNumber` 等字段属于页面运行时状态，由 Chat Store 独立维护，不属于 iframe query 初始化参数。

### 两次 `/chat/welcome` 的职责

当前初始化流程会调用两次 `/chat/welcome`，两次请求的职责不同：

1. 路由阶段由 `setH5Config()` 预取机器人展示配置和语言配置。
2. 聊天页面获取历史会话并确定 `dialogue_id` 后，由 `createChat()` 创建或恢复完整会话，同时初始化欢迎语、会话变量和 IM 连接。

因此不要仅根据接口地址相同合并这两次请求。调整初始化流程前，需要先确认后端对配置预取和会话初始化的接口约定。

### 添加新的初始化参数

例如需要增加 `customer_level`：

1. 在 `src/types/chat.ts` 的 `ChatInitParams` 中声明：

```ts
customer_level: string
```

2. 在 `src/utils/chat-init.ts` 的 `parseChatInitParams()` 中统一解析：

```ts
customer_level: String(query.customer_level || '')
```

3. 根据字段用途继续处理：

   - 仅供页面读取：在对应页面或 Store 中使用 `data.customer_level`。
   - 需要保存为用户状态：在 Chat Store 的 `user` 中增加字段，并在 `setH5Config()`、`createChat()` 中赋值。
   - 需要提交后端：在对应 `/chat/welcome` 请求参数中增加字段，并先确认后端接口支持。

4. 在 `chat-ai-sdk` 接入方的 `data-json` 中传入字段：

```js
JSON.stringify({
  robot_key: 'YOUR_ROBOT_KEY',
  openid: 'user_10001',
  customer_level: 'vip'
})
```

不需要再分别修改路由守卫和聊天页面的 query 解析逻辑；两处均复用 `parseChatInitParams()`。初始化参数会出现在 iframe URL 中，敏感信息应改用经过安全设计的通信方式传递。

This template should help get you started developing with Vue 3 in Vite.

## Recommended IDE Setup

[VSCode](https://code.visualstudio.com/) + [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Type Support for `.vue` Imports in TS

TypeScript cannot handle type information for `.vue` imports by default, so we replace the `tsc` CLI with `vue-tsc` for type checking. In editors, we need [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) to make the TypeScript language service aware of `.vue` types.

## Customize configuration

See [Vite Configuration Reference](https://vitejs.dev/config/).

## Project Setup

```sh
npm install
```

### Compile and Hot-Reload for Development

```sh
npm run dev
```

### Type-Check, Compile and Minify for Production

```sh
npm run build
```

### Run Unit Tests with [Vitest](https://vitest.dev/)

```sh
npm run test:unit
```

### Lint with [ESLint](https://eslint.org/)

```sh
npm run lint
```
