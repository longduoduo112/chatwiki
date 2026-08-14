<template>
  <a-modal
    class="create-book-skill-modal"
    :open="visible"
    :title="isUpdate ? t('modal_update_document') : t('modal_upload_document')"
    :width="640"
    :maskClosable="false"
    :destroyOnClose="false"
    @cancel="handleCancel"
  >
    <a-form class="create-form" layout="vertical">
      <a-form-item :label="t('label_custom_prompt')">
        <a-textarea
          v-model:value="formState.custom_prompt"
          :auto-size="{ minRows: 3, maxRows: 6 }"
          :placeholder="t('placeholder_custom_prompt')"
          :disabled="submitLoading"
        />
      </a-form-item>

      <a-form-item required :label="t('label_generate_model')">
        <ModelSelect
          modelType="LLM"
          v-model:modeName="formState.use_model"
          v-model:modeId="formState.model_config_id"
          :placeholder="t('placeholder_select_model')"
        />
      </a-form-item>

      <a-form-item required :label="t('label_temperature')">
        <a-input-number
          v-model:value="formState.temperature"
          :min="0"
          :max="2"
          :step="0.1"
          :disabled="submitLoading"
        />
      </a-form-item>

      <a-form-item required :label="t('label_max_token')">
        <a-input-number
          v-model:value="formState.max_token"
          :min="1"
          :precision="0"
          :disabled="submitLoading"
        />
      </a-form-item>
    </a-form>

    <a-upload-dragger
      class="book-upload-dragger"
      :file-list="fileList"
      :show-upload-list="false"
      :multiple="true"
      :max-count="20"
      accept=".txt,.docx,.md,.pdf"
      :disabled="submitLoading"
      :beforeUpload="handleBeforeUpload"
    >
      <div class="upload-icon">
        <InboxOutlined />
      </div>
      <div class="upload-title">{{ t('upload_title') }}</div>
      <div class="upload-hint">
        {{ t('upload_limit_hint') }}<br />
        {{ t('upload_type_hint') }}
      </div>
    </a-upload-dragger>

    <div v-if="fileList.length" class="file-list">
      <div v-for="file in fileList" :key="file.uid || file.name" class="file-row">
        <div class="file-main">
          <LoadingOutlined v-if="submitLoading" class="file-loading" />
          <FileTextOutlined v-else class="file-icon" />
          <div class="file-info">
            <div class="file-name">{{ file.name }}</div>
            <a-progress
              v-if="submitLoading"
              class="file-progress"
              :percent="80"
              :show-info="false"
              size="small"
            />
          </div>
        </div>
        <DeleteOutlined class="delete-icon" @click="handleRemoveFile(file)" />
      </div>
    </div>

    <div v-if="hasPdfFiles" class="online-ocr-section">
      <div class="online-ocr-label">{{ t('label_online_ocr') }}</div>
      <div class="online-ocr-switch-row">
        <a-switch v-model:checked="formState.online_ocr" :disabled="submitLoading" />
        <span>{{ formState.online_ocr ? t('online_ocr_enabled') : t('online_ocr_disabled') }}</span>
      </div>
      <div class="online-ocr-tip">
        <InfoCircleOutlined />
        <span>{{ t('online_ocr_format_tip') }}</span>
      </div>
      <div class="online-ocr-tip">
        <InfoCircleOutlined />
        <span>
          {{ t('online_ocr_points_tip') }}
          <span class="points-balance">{{ t('online_ocr_points_balance') }}</span>
          <a
            class="points-link"
            href="https://cloud.chatwiki.com/#/user/model"
            target="_blank"
            rel="noopener noreferrer"
          >{{ t('online_ocr_points_view') }}</a>
          <a
            class="purchase-link"
            href="https://cloud.chatwiki.com/#/user/model?open_points=1"
            target="_blank"
            rel="noopener noreferrer"
          >{{ t('online_ocr_purchase') }}</a>
        </span>
      </div>
    </div>

    <template #footer>
      <a-button :disabled="submitLoading" @click="handleCancel">{{ t('btn_cancel') }}</a-button>
      <a-button type="primary" :loading="submitLoading" @click="handleConfirm">{{ t('btn_confirm') }}</a-button>
    </template>
  </a-modal>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import {
  DeleteOutlined,
  FileTextOutlined,
  InboxOutlined,
  InfoCircleOutlined,
  LoadingOutlined
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { useI18n } from '@/hooks/web/useI18n'
import ModelSelect from '@/components/model-select/model-select.vue'
import { createDocToSkillTask, updateDocToSkillTask } from '@/api/clawbot'

const MAX_FILE_COUNT = 20
const MAX_FILE_SIZE = 100 * 1024 * 1024
const ACCEPT_EXTS = ['txt', 'docx', 'md', 'pdf']
const { t } = useI18n('views.clawbot.skill-generate-tool.index')

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  task: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:visible', 'confirm'])

const fileList = ref([])
const submitLoading = ref(false)
const isUpdate = computed(() => Number(props.task?.id) > 0)
const hasPdfFiles = computed(() => fileList.value.some((file) => getFileExt(file?.name) === 'pdf'))
const formState = reactive({
  custom_prompt: '',
  model_config_id: '',
  use_model: '',
  temperature: 1,
  max_token: 32768,
  online_ocr: false
})

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      resetForm()
    }
  }
)

const resetForm = () => {
  const task = isUpdate.value ? props.task || {} : {}
  formState.custom_prompt = task.custom_prompt || ''
  formState.model_config_id = task.model_config_id || ''
  formState.use_model = task.use_model || ''
  formState.temperature = task.temperature ?? 1
  formState.max_token = task.max_token ?? 32768
  formState.online_ocr = false
  fileList.value = []
  submitLoading.value = false
}

const getFileExt = (fileName = '') => {
  return fileName.split('.').pop()?.toLowerCase() || ''
}

const validateFile = (file) => {
  const ext = getFileExt(file.name)
  if (!ACCEPT_EXTS.includes(ext)) {
    message.error(t('msg_unsupported_book_format'))
    return false
  }

  if (file.size > MAX_FILE_SIZE) {
    message.error(t('msg_file_too_large'))
    return false
  }

  if (file.size <= 0) {
    message.error(t('msg_empty_file'))
    return false
  }

  if (fileList.value.length >= MAX_FILE_COUNT) {
    message.error(t('msg_max_file_count'))
    return false
  }

  return true
}

const handleBeforeUpload = (file) => {
  if (!validateFile(file)) {
    return false
  }
  fileList.value = [...fileList.value, file]
  return false
}

const handleRemoveFile = (file) => {
  if (submitLoading.value) {
    return
  }
  fileList.value = fileList.value.filter((item) => item.uid !== file.uid)
}

const validateForm = () => {
  if (!formState.model_config_id || !formState.use_model) {
    message.error(t('msg_select_model'))
    return false
  }
  if (!fileList.value.length) {
    message.error(t('msg_upload_document'))
    return false
  }
  const temperature = Number(formState.temperature)
  if (!Number.isFinite(temperature) || temperature < 0 || temperature > 2) {
    message.error(t('msg_invalid_temperature'))
    return false
  }
  const maxToken = Number(formState.max_token)
  if (!Number.isInteger(maxToken) || maxToken <= 0) {
    message.error(t('msg_invalid_doc_max_token'))
    return false
  }
  return true
}

const handleConfirm = async () => {
  if (!validateForm()) {
    return
  }

  submitLoading.value = true
  try {
    const formData = new FormData()
    formData.append('custom_prompt', formState.custom_prompt.trim())
    formData.append('model_config_id', formState.model_config_id)
    formData.append('use_model', formState.use_model)
    formData.append('temperature', String(formState.temperature))
    formData.append('max_token', String(formState.max_token))
    formData.append('online_ocr', String(hasPdfFiles.value && formState.online_ocr))
    if (isUpdate.value) {
      formData.append('id', String(props.task.id))
    }
    fileList.value.forEach((file) => {
      formData.append('files', file)
    })

    const res = isUpdate.value
      ? await updateDocToSkillTask(formData)
      : await createDocToSkillTask(formData)
    if (res && (res.res === 0 || res.code === 0)) {
      message.success(t(isUpdate.value ? 'msg_doc_update_created' : 'msg_task_created'))
      emit('confirm', res.data)
      emit('update:visible', false)
    } else {
      message.error(res?.msg || t(isUpdate.value ? 'msg_doc_update_create_failed' : 'msg_task_create_failed'))
    }
  } catch (err) {
    console.error(isUpdate.value ? '更新Book转Skill任务失败' : '创建Book转Skill任务失败', err)
  } finally {
    submitLoading.value = false
  }
}

const handleCancel = () => {
  if (submitLoading.value) {
    return
  }
  emit('update:visible', false)
}
</script>

<style lang="less" scoped>
.create-form {
  padding-top: 10px;

  :deep(.ant-form-item) {
    margin-bottom: 16px;
  }

  :deep(.ant-form-item-label) {
    padding-bottom: 6px;
  }

  :deep(.ant-form-item-label > label) {
    color: #262626;
    font-size: 14px;
    line-height: 22px;
  }

  :deep(.ant-input),
  :deep(.ant-input-number),
  :deep(.ant-select-selector) {
    border-radius: 6px;
  }

  :deep(.ant-input-number) {
    width: 100%;
  }
}

.book-upload-dragger {
  display: block;

  :deep(.ant-upload-drag) {
    border: 1px dashed #d6dce6;
    border-radius: 6px;
    background: #f7f9fc;
    padding: 20px 16px 18px;
  }
}

.upload-icon {
  color: #2475fc;
  font-size: 38px;
  line-height: 1;
}

.upload-title {
  margin-top: 16px;
  color: #262626;
  font-size: 14px;
  line-height: 22px;
}

.upload-hint {
  margin-top: 4px;
  color: #8c8c8c;
  font-size: 14px;
  line-height: 22px;
}

.file-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 8px;
}

.online-ocr-section {
  margin-top: 18px;
}

.online-ocr-label {
  color: #262626;
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
}

.online-ocr-switch-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
  color: #8c8c8c;
  font-size: 14px;
}

.online-ocr-tip {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-top: 8px;
  padding: 8px 12px;
  border: 1px solid #ffd666;
  border-radius: 6px;
  background: #fffbe6;
  color: #fa8c16;
  font-size: 13px;
  line-height: 20px;

  > :first-child {
    margin-top: 3px;
    flex-shrink: 0;
  }
}

.points-balance,
.purchase-link {
  margin-left: 4px;
}

.points-link,
.purchase-link {
  color: #1677ff;
}

.file-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 50px;
  padding: 8px 12px;
  border: 1px dashed #d6dce6;
  border-radius: 6px;
  background: #fff;
}

.file-main {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
}

.file-icon,
.file-loading {
  color: #2475fc;
  font-size: 16px;
  flex-shrink: 0;
}

.file-info {
  flex: 1;
  min-width: 0;
}

.file-name {
  color: #262626;
  font-size: 14px;
  line-height: 22px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-progress {
  margin-top: 3px;
  line-height: 1;
}

.delete-icon {
  color: #595959;
  cursor: pointer;
  font-size: 16px;
  flex-shrink: 0;

  &:hover {
    color: #ff4d4f;
  }
}

:deep(.ant-modal-content) {
  border-radius: 14px;
}

:deep(.ant-modal-header) {
  margin-bottom: 12px;
}

:deep(.ant-modal-title) {
  color: #262626;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
}

:deep(.ant-modal-footer) {
  margin-top: 24px;
}
</style>
