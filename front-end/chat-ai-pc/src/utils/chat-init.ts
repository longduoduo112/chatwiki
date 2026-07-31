import type { LocationQuery } from 'vue-router'
import type { ChatInitParams } from '@/types/chat'
import { getOpenid } from '@/utils/index'

// 路由守卫和聊天页面共用同一解析入口，新增初始化参数时在此统一处理默认值和类型转换。
export function parseChatInitParams(query: LocationQuery): ChatInitParams {
  return {
    openid: String(query.openid || getOpenid()),
    robot_key: String(query.robot_key || ''),
    avatar: String(query.avatar || ''),
    name: String(query.name || ''),
    nickname: String(query.nickname || ''),
    dialogue_id: Number(query.dialogue_id) || 0
  }
}
