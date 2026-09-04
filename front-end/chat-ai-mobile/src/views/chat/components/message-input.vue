<style lang="less" scoped>
.message-input-box {
  width: calc(100% - 24px);
  max-width: 900px;
  box-sizing: border-box;
  margin: 12px auto;
  padding: 7px 12px;
  overflow: hidden;
  background: #fff;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  box-shadow: 0 4px 8px 0 rgba(0, 0, 0, 0.08);
  transition: all 0.2s;

  &.is-active {
    border-color: #5694fc;
    box-shadow: 0 4px 16px 0 rgba(0, 149, 255, 0.18);
  }

  .message-input-body {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .message-input {
    display: flex;
    position: relative;
    width: 100%;
    overflow: hidden;
    padding: 0;

    .text-input {
      line-height: 24px;
      height: 24px;
      width: 100%;
      padding: 0;
      font-size: 16px;
      font-weight: 400;
      color: rgb(26, 26, 26);
      background: none;
      outline: none;
      resize: none;
      border: none;
      transition: height 0.1s ease-in-out;
      overflow: hidden;
      white-space: pre-wrap; /* 保持内容的换行，并允许自动换行 */

      &::placeholder {
        font-size: 16px;
        font-weight: 400;
        color: #8c8c8c;
      }
    }

    /* 滚动条样式 */
    .text-input::-webkit-scrollbar {
      width: 4px; /*  设置纵轴（y轴）轴滚动条 */
      height: 4px; /*  设置横轴（x轴）轴滚动条 */
    }
    /* 滚动条滑块（里面小方块） */
    .text-input::-webkit-scrollbar-thumb {
      cursor: pointer;
      border-radius: 5px;
      background: transparent;
    }
    /* 滚动条轨道 */
    .text-input::-webkit-scrollbar-track {
      border-radius: 5px;
      background: transparent;
    }

    /* hover时显色 */
    .text-input:hover::-webkit-scrollbar-thumb {
      background: rgba(0, 0, 0, 0.2);
    }
    .text-input:hover::-webkit-scrollbar-track {
      background: rgba(0, 0, 0, 0.1);
    }
  }

  .message-action {
    display: flex;
    align-items: center;
    justify-content: space-between;

    .send-btn {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 32px;
      height: 32px;
      padding: 0;
      margin: 0 0 0 auto;
      font-size: 28px;
      border: none;
      outline: none;
      background: none;
      transition: all 0.2s;
      cursor: pointer;
      color: #2475fc;

      &:hover {
        opacity: 0.8;
      }
      &:disabled {
        opacity: 0.3;
      }

      .send-pause {
        font-size: 28px;
      }
    }

    .file-action {
      position: relative;
      display: flex;
      align-items: center;
      height: 32px;
      padding: 0;

      .action-btn {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 28px;
        height: 28px;
        padding: 6px;
        font-size: 16px;
        color: #595959;
        background: #fff;
        border-radius: 6px;
        box-sizing: border-box;
        cursor: pointer;

        &:hover {
          color: #2475fc;
          background: #e4e6eb;
        }
      }

      .file-number {
        position: absolute;
        left: 16px;
        top: -8px;
        width: 16px;
        height: 16px;
        border-radius: 50%;
        background: #f00;
        color: #fff;
        font-size: 12px;
        font-weight: 400;
        display: flex;
        align-items: center;
        justify-content: center;

        &.big {
          width: auto;
          padding: 0 4px;
          border-radius: 12px;
        }
      }
    }
  }
}
</style>

<template>
  <div class="message-input-box" :class="{ 'is-active': isFocus || props.value || fileList.length }">
    <FileToolbar :file-list="props.fileList" @delete="deleteFile" v-if="props.fileList.length > 0" />
    <div class="message-input-body">
        <div class="message-input">
          <textarea
            ref="messageTextarea"
            :style="{ height: inputHeight + 'px' }"
            class="text-input"
            :value="props.value"
            :placeholder="t('ph_input_message')"
            @change="onChange"
            @input="onInput"
            @keydown.enter="handleEnter"
            @focus="onFocus"
            @blur="onBlur"
          ></textarea>
        </div>

        <div class="message-action">
          <div class="file-action" v-if="props.showUpload">
            <span class="file-number" :class="{ big: fileList.length > 9 }" v-if="fileList.length > 0">{{ fileList.length }}</span>
            <svg-icon class="action-btn select-file" name="circularNeedle" @click="openFileDialog"></svg-icon>
          </div>

          <button
            class="send-btn"
            :class="{ loading: props.loading }"
            :disabled="disabled"
            :title="props.loading ? '停止发送' : ''"
            @click="props.loading ? stopMessage() : sendMessage()"
          >
            <svg-icon class="send-pause" name="send-pause" v-if="props.loading"></svg-icon>
            <svg-icon name="send-message" width="28px" height="28px" v-else></svg-icon>
          </button>
        </div>
    </div>
  </div>
</template>

<script setup>
import { ref, toRefs, computed, watch, nextTick  } from 'vue'
import calcTextareaHeight from '@/utils/calcTextareaHeight'
import { useChatStore } from '@/stores/modules/chat'
import { useUserStore } from '@/stores/modules/user'
import { showToast } from 'vant'
import { useI18n } from '@/hooks/web/useI18n'
import { useUpload } from '@/hooks/web/useUpload.js'
import FileToolbar from './file-toolbar.vue'

const chatStore = useChatStore()
const userStore = useUserStore()
const { robot } = chatStore
import { checkChatRequestPermission } from '@/api/robot/index'

const emit = defineEmits(['update:value', 'send', 'stop', 'showLogin', 'update:fileList'])

const props = defineProps({
  value: {
    type: String,
    default: ''
  },
  loading: {
    type: Boolean,
    default: false
  },
  fileList: {
    type: Array,
    default: () => []
  },
  showUpload: {
    type: Boolean,
    default: false
  },
})

const { t } = useI18n('views.chat.components.message-input')

const { fileList } = toRefs(props)
const messageTextarea = ref(null)
const isFocus = ref(false)

const disabled = computed(() => {
  if (props.loading) {
    return false
  }

  if(props.fileList.length > 0){
    return false
  }

  return props.value.trim().length === 0
})


const { openFileDialog } = useUpload({
  limit: 10,
  maxSize: 10,
  category: 'chat_image',
  fileList: fileList,
  multiple: true,
  accept: 'image/bmp,image/jpeg,image/png,image/tiff,image/heic,image/gif,image/webp',
  extraData: {
    robot_key: robot.robot_key,
    openid: robot.openid
  }
})

const deleteFile = (index) => {
  const newFileList = props.fileList.filter((_, i) => i !== index);
  emit('update:fileList', newFileList);
}

const onChange = (event) => {
  emit('update:value', event.target.value)
}

const inputHeight = ref(24)
const inputMaxHeight = 5 * 24
const setInputHeight = (value) => {
  if (!value) {
    // 如果值为空，重置为初始高度
    inputHeight.value = 24
    if (messageTextarea.value) {
      messageTextarea.value.style.overflow = 'hidden'
    }
  } else {
    // 如果值不为空，重新计算高度
    nextTick(() => {
      let newHeight = calcTextareaHeight(messageTextarea.value).height || 24
      newHeight = parseInt(newHeight)

      if (newHeight >= inputMaxHeight) {
        inputHeight.value = inputMaxHeight
        if (messageTextarea.value) {
          messageTextarea.value.style.overflow = 'auto'
        }
      } else {
        inputHeight.value = newHeight
        if (messageTextarea.value) {
          messageTextarea.value.style.overflow = 'hidden'
        }
      }
    })
  }
}

const onInput = (event) => {
  onChange(event)

  // setInputHeight(event.target.value)
}

// 监听 props.value 变化，当值改变时调整高度
watch(() => props.value, (newValue) => {
  setInputHeight(newValue)
}, { immediate: false })

const sendMessage = async () => {
  emit('send', props.value)
}

const stopMessage = () => {
  emit('stop')
}

const handleEnter = (event) => {
  const target = event.target
  
  if (event.shiftKey) {
    emit('update:value', event.target.value)
  }else{
    if (!event.target.value) {
      return
    }

    event.preventDefault()
    event.stopPropagation()

    sendMessage()
  }
}

const onFocus = () => {
  isFocus.value = true
}

const onBlur = () => {
  isFocus.value = false
}

const handleSetValue = (data) => {
  emit('update:value', data)
}


defineExpose({
  handleSetValue,
  sendMessage
})
</script>
