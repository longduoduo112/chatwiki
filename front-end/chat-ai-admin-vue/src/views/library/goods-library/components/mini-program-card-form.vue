<template>
  <div class="mini-program-card-form">
    <div class="card-form-header">
      <span>{{ t('goods_card.mini_program') }}</span>
      <a-button
        type="text"
        class="remove-card-btn"
        :aria-label="t('goods_card.remove')"
        @click="emit('remove')"
      >
        <CloseOutlined />
      </a-button>
    </div>

    <div class="card-form-body">
      <a-form :label-col="{ span: 5 }" :wrapper-col="{ span: 19 }">
        <a-form-item :label="t('goods_card.appid_label')">
          <a-input
            :value="modelValue.appid"
            :maxlength="100"
            :placeholder="t('goods_card.appid_placeholder')"
            @update:value="(value) => updateField('appid', value)"
          />
        </a-form-item>

        <a-form-item :label="t('goods_card.path_label')">
          <a-input
            :value="modelValue.path"
            :maxlength="500"
            :placeholder="t('goods_card.path_placeholder')"
            @update:value="(value) => updateField('path', value)"
          />
        </a-form-item>

        <a-form-item :label="t('goods_card.title_label')">
          <a-input
            :value="modelValue.title"
            :maxlength="255"
            :placeholder="t('goods_card.title_placeholder')"
            @update:value="(value) => updateField('title', value)"
          />
        </a-form-item>

        <a-form-item :label="t('goods_card.image_label')" class="card-image-item">
          <a-upload
            v-model:fileList="fileList"
            list-type="picture-card"
            :max-count="1"
            accept="image/*"
            :before-upload="handleBeforeUpload"
            :custom-request="handleCustomUpload"
            @change="handleUploadChange"
            @preview="handlePreview"
          >
            <div v-if="fileList.length < 1" class="upload-trigger">
              <PlusOutlined class="upload-icon" />
              <div class="upload-text">{{ t('goods_modal.upload_action') }}</div>
            </div>
          </a-upload>
          <div class="upload-tip">{{ t('goods_card.image_hint') }}</div>
        </a-form-item>
      </a-form>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { message, Upload } from 'ant-design-vue'
import { CloseOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { api as viewerApi } from 'v-viewer'
import { uploadGoodsImage } from '@/api/goods-library'
import { useI18n } from '@/hooks/web/useI18n'
import { generateRandomId } from '@/utils/index'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({
      appid: '',
      path: '',
      title: '',
      image: ''
    })
  }
})

const emit = defineEmits(['update:modelValue', 'remove'])

const { t } = useI18n('views.library.goods-library.index')
const fileList = ref([])

const updateField = (field, value) => {
  emit('update:modelValue', {
    ...props.modelValue,
    [field]: value
  })
}

const syncImageFile = (image) => {
  if (!image) {
    fileList.value = []
    return
  }

  const currentImage = fileList.value[0]?.url || fileList.value[0]?.thumbUrl || ''
  if (currentImage === image) return

  fileList.value = [{
    uid: generateRandomId(8),
    name: image.split('/').pop() || 'card-image',
    url: image,
    thumbUrl: image,
    status: 'done'
  }]
}

watch(() => props.modelValue.image, syncImageFile, { immediate: true })

const handleBeforeUpload = (file) => {
  if (file.size > 1024 * 1024) {
    message.error(t('goods_card.image_size_limit'))
    return Upload.LIST_IGNORE
  }

  return true
}

const handleCustomUpload = async ({ file, onSuccess, onError }) => {
  try {
    const res = await uploadGoodsImage({ file })
    const link = res?.data?.link || ''
    onSuccess({ link }, file)
  } catch (error) {
    onError(error)
  }
}

const handleUploadChange = (info) => {
  const nextList = info.fileList.slice(0, 1)
  const currentFile = nextList[0]

  if (currentFile?.status === 'done' && currentFile.response?.link) {
    currentFile.url = currentFile.response.link
    currentFile.thumbUrl = currentFile.response.link
  }

  fileList.value = nextList
  updateField('image', currentFile?.url || currentFile?.thumbUrl || '')
}

const handlePreview = (file) => {
  const image = file.url || file.thumbUrl
  if (!image) return

  viewerApi({
    images: [image],
    options: {
      toolbar: true,
      title: false,
      movable: true,
      zoomable: true,
      rotatable: true,
      scalable: true
    }
  })
}
</script>

<style lang="less" scoped>
.mini-program-card-form {
  overflow: hidden;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  background: #fff;
}

.card-form-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 40px;
  padding: 0 12px 0 16px;
  border-bottom: 1px solid #f0f0f0;
  background: #fafafa;
  color: #262626;
  font-size: 14px;
  font-weight: 500;
}

.remove-card-btn {
  width: 24px;
  height: 24px;
  padding: 0;
  color: #bfbfbf;
}

.card-form-body {
  padding: 16px 16px 4px;

  :deep(.ant-form-item) {
    margin-bottom: 12px;
  }

  :deep(.ant-form-item-label > label) {
    color: #262626;
  }
}

.card-image-item {
  :deep(.ant-upload-wrapper.ant-upload-picture-card-wrapper .ant-upload.ant-upload-select) {
    width: 100px;
    height: 100px;
    margin-bottom: 0;
  }
}

.upload-trigger {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}

.upload-icon {
  color: #8c8c8c;
  font-size: 22px;
}

.upload-text {
  margin-top: 8px;
  color: #8c8c8c;
  font-size: 13px;
}

.upload-tip {
  margin-top: 6px;
  color: #8c8c8c;
  font-size: 13px;
}
</style>
