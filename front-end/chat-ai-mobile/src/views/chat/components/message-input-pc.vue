<template>
  <div class="message-input-wrapper" :class="{ 'is-set': props.value || fileList.length }">
    <FileToolbar :file-list="fileList" @delete="deleteFile" v-if="fileList.length > 0" />
    <div class="message-input-box">
      <ATextarea
        class="message-input"
        :value="props.value"
        :auto-size="{ minRows: 1, maxRows: 5 }"
        :placeholder="t('ph_input_message_with_shift')"
        @change="onChange"
        @keydown="handleKeydown"
      />
    </div>

    <div class="message-action">
      <div class="select-file-btn" @click="openFileDialog" v-if="props.showUpload">
        <svg-icon class="select-file-icon" name="circularNeedle"></svg-icon>
        <span class="file-number" :class="{ big: fileList.length > 9 }" v-if="fileList.length > 0">{{ fileList.length }}</span>
      </div>

      <ATooltip v-if="props.loading" title="停止发送">
        <button class="send-msg-btn loading" @click="stopMessage">
          <svg-icon class="send-pause" name="send-pause" />
        </button>
      </ATooltip>
      <button
        v-else
        class="send-msg-btn"
        :disabled="disabled"
        @click="sendMessage"
      >
        <svg-icon class="paper-airplane" name="send-message" width="28px" height="28px" />
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, toRefs, computed } from 'vue'
import { useChatStore } from '@/stores/modules/chat'
import { useUserStore } from '@/stores/modules/user'
import { Textarea as ATextarea, Tooltip as ATooltip } from 'ant-design-vue'
import { showToast } from 'vant'
import { useI18n } from '@/hooks/web/useI18n'
import { useUpload } from '@/hooks/web/useUpload.js'
import { checkChatRequestPermission } from '@/api/robot/index'
import FileToolbar from './file-toolbar.vue'

const chatStore = useChatStore()
const userStore = useUserStore()
const { robot } = chatStore


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

const { fileList } = toRefs(props)

const { t } = useI18n('views.chat.components.message-input-pc')

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

const disabled = computed(() => {
  if(props.fileList.length > 0){
    return false
  }

  return props.loading || props.value.trim().length === 0
})

const onChange = (event) => {
  emit('update:value', event.target.value)
}

const sendMessage = async () => {
  emit('send', props.value)
}

const stopMessage = () => {
  emit('stop')
}

const handleKeydown = (event) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    if (!event.target.value) {
      return
    }
    event.preventDefault()
    event.stopPropagation()
    sendMessage()
  } else if (event.key === 'Enter' && event.shiftKey) {
    emit('update:value', event.target.value)
  }
}

const handleSetValue = (data) => {
  emit('update:value', data)
}

defineExpose({
  handleSetValue,
  sendMessage
})
</script>


<style lang="less" scoped>
.message-input-wrapper {
  position: relative;
  max-width: 900px;
  margin: 0 auto;
  padding: 10px 12px;
  border-radius: 12px;
  border: 2px solid #e5e7eb;
  overflow: hidden;
  background-color: #fff;
  box-shadow: 0 4px 8px 0 rgba(0, 0, 0, 0.08);
  transition: all 0.2s;
  box-sizing: border-box;

  &.is-set,
  &:focus-within {
    border-color: #5694fc;
    box-shadow: 0 4px 16px 0 rgba(0, 149, 255, 0.18);
  }

  .message-input-box {
    padding: 0;
    margin-bottom: 16px;
  }

  .message-input {
    width: 100%;
    padding: 0;
    color: #262626;
    font-size: 16px;
    line-height: 24px;
    border: none !important;
    outline: none !important;
    resize: none !important;
    box-shadow: none !important;

    &::placeholder {
      color: #8c8c8c;
      font-size: 16px;
    }
  }

  .message-action {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0;
  }

  .send-msg-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    padding: 0;
    margin-left: auto;
    font-size: 14px;
    font-weight: 400;
    border-radius: 6px;
    border: none;
    cursor: pointer;
    transition: all 0.2s;
    color: #2475fc;
    background: none;

    &:hover {
      opacity: 0.8;
    }
    &:disabled {
      opacity: 0.3;
    }
    .paper-airplane {
      font-size: 28px;
    }

    &.loading {
      color: #2475fc;
    }
    .send-pause {
      font-size: 28px;
    }
    
  }

  .select-file-btn {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    padding: 0;
    margin-right: auto;
    border-radius: 6px;
    border: none;
    background: #fff;
    cursor: pointer;
    transition: all 0.2s;
    border: 0;

    .is-set &,
    .message-input-wrapper:focus-within & {
      background: #e4e6eb;
    }

    &:hover {
      background: #e4e6eb;
    }

    .select-file-icon {
      font-size: 16px;
      color: #595959;
    }

    .file-number{
      position: absolute;
      right: -8px;
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

      &.big{
        width: auto;
        padding: 0 4px;
        border-radius: 12px;
      }
    }
  }
}
</style>
