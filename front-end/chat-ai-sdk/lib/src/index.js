
import './style.css'
import AiChatWidget from './ai-chat'

export function init() {
  let config = {
    iframeSrc: import.meta.env.VITE_AI_CHAT_BASE_URL + '/#/chat',
    remote: '',
    params: {},
    showFloatButton: true
  };

  const sdkEl = document.getElementById("ai_chat_js")

  if(sdkEl){
    let params = sdkEl.getAttribute("data-json")
    let origin = new URL(sdkEl.src).origin

    if(import.meta.env.DEV){
      // 开发者模式iframe地址使用本地地址
      config.iframeSrc = import.meta.env.VITE_AI_CHAT_BASE_URL + '/#/chat'
    }else{
      config.iframeSrc = origin + '/web/#/chat'
    }
    
    try{
      const parsedParams = JSON.parse(params) || {}

      if (Object.prototype.hasOwnProperty.call(parsedParams, 'show_float_button')) {
        const showFloatButton = parsedParams.show_float_button
        config.showFloatButton = showFloatButton !== false && showFloatButton !== 0 && showFloatButton !== '0'
        delete parsedParams.show_float_button
      }

      config.params = parsedParams
    } catch (error) {
      console.error('Failed to stringify data:', error);
      return;
    }
  }

  return AiChatWidget.init(config)
}

