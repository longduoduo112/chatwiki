// 这里只描述 iframe query 带入的会话初始化参数；isOpen、unreadNumber 等状态由 Chat Store 独立维护。
export interface ChatInitParams {
  openid: string
  robot_key: string
  avatar: string
  name: string
  nickname: string
  dialogue_id: number
}
