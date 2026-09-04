<style lang="less" scoped>
.message-input-box {
  margin: 12px;
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

      .line {
        width: 1px;
        height: 14px;
        margin: 0 8px;
        background: #D9D9D9;
      }

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

      .show-file{
        color: #2475FC;
      }

      .hide-file{
        color: #595959;
      }

      .file-number{
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

        &.big{
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
  <div class="message-input-box" :class="{ 'is-active': isFocus || value || fileList.length }">
    <FileToolbar :file-list="props.fileList" @delete="deleteFile" v-if="props.fileList.length > 0 && showFiletoolbar" />
    <div class="message-input-body">
      <div class="message-input">
        <AutoSizeTextarea
          :value="value"
          @change="onChange"
          @focus="onFocus"
          @blur="onBlur"
          @enter="sendMessage"
        />
      </div>
      <div class="message-action">
        <div class="file-action" v-if="props.showUpload">
          <span class="file-number" :class="{ big: fileList.length > 9 }" v-if="fileList.length > 0">{{ fileList.length }}</span>
          <svg-icon class="action-btn select-file" name="circularNeedle" @click="openFileDialog"></svg-icon>
          <i class="line" v-if="fileList.length > 0"></i>
          <Tippy
            :content="showFiletoolbar ? t('msg_hide_images') : t('msg_show_images')"
            placement="top"
            v-if="fileList.length > 0"
          >
            <svg-icon class="action-btn show-file" name="eye-open" v-if="showFiletoolbar" @click="showFiletoolbar = false"></svg-icon>
            <svg-icon class="action-btn hide-file" name="eye-close" v-else @click="showFiletoolbar = true"></svg-icon>
          </Tippy>
        </div>

        <Tippy content="停止发送" placement="top" v-if="props.loading">
          <button class="send-btn" @click="stopMessage">
            <svg-icon class="send-pause" name="send-pause" />
          </button>
        </Tippy>
        <button class="send-btn" v-else @click="sendMessage" :disabled="disabled">
          <svg-icon name="send-message" width="28px" height="28px" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, toRefs, computed } from 'vue'
import { useI18n } from '@/hooks/web/useI18n'
import { useUpload } from '@/hooks/web/useUpload.js'
import { useChatStore } from '@/stores/modules/chat'
import AutoSizeTextarea from './auto-size-textarea.vue'
import FileToolbar from './file-toolbar.vue'
import { Tippy } from 'vue-tippy'

const { t } = useI18n('views.chat.components.message-input')


const emit = defineEmits(['update:value', 'send', 'stop', 'update:fileList'])

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

const chatStore = useChatStore()

const { robot } = chatStore

const isFocus = ref(false)
const { fileList } = toRefs(props)
const showFiletoolbar = ref(true)

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
  if (props.loading) {
    return false
  }

  if(props.fileList.length > 0){
    return false
  }

  return props.value.trim().length === 0
})

const onChange = (val: string) => {
  emit('update:value', val)
}

const sendMessage = () => {
  emit('send', props.value)
}

const stopMessage = () => {
  emit('stop')
}

const onFocus = () => {
  isFocus.value = true
}

const onBlur = () => {
  isFocus.value = false
}
</script>
