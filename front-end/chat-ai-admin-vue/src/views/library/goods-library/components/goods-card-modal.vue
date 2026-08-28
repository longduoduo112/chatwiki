<template>
  <a-modal
    v-model:open="open"
    :title="t('goods_card.modal_title')"
    :ok-text="t('field_editor.save')"
    :cancel-text="t('field_editor.cancel')"
    :width="500"
    :destroy-on-close="true"
    :confirm-loading="submitting"
    @ok="handleSave"
    @cancel="close"
  >
    <MiniProgramCardForm
      v-model="cardModel"
      @remove="resetCard"
    />
  </a-modal>
</template>

<script setup>
import { ref } from 'vue'
import { message } from 'ant-design-vue'
import { useI18n } from '@/hooks/web/useI18n'
import MiniProgramCardForm from './mini-program-card-form.vue'

const { t } = useI18n('views.library.goods-library.index')
const emit = defineEmits(['save'])
const open = ref(false)
const submitting = ref(false)
const currentRow = ref(null)

const createEmptyCard = () => ({
  appid: '',
  path: '',
  title: '',
  image: ''
})

const cardModel = ref(createEmptyCard())

const resetCard = () => {
  cardModel.value = createEmptyCard()
}

const normalizeCard = (card) => {
  const source = card || {}
  return {
    appid: String(source.appid || '').trim(),
    path: String(source.path || '').trim(),
    title: String(source.title || '').trim(),
    image: String(source.image || '').trim()
  }
}

const hasCardValue = (card) => Object.values(card).some((value) => Boolean(value))

const close = () => {
  open.value = false
  submitting.value = false
  currentRow.value = null
  resetCard()
}

const show = (row = {}) => {
  currentRow.value = row
  cardModel.value = normalizeCard(row.goods_wechat_card)
  submitting.value = false
  open.value = true
}

const handleSave = () => {
  if (submitting.value) return

  const card = normalizeCard(cardModel.value)
  if (hasCardValue(card) && Object.values(card).some((value) => !value)) {
    message.error(t('goods_card.required_fields'))
    return
  }
  submitting.value = true
  emit('save', {
    row: currentRow.value,
    goods_wechat_card: hasCardValue(card) ? card : {}
  }, {
    close,
    setSubmitting: (value) => {
      submitting.value = value
    }
  })
}

defineExpose({
  show
})
</script>
