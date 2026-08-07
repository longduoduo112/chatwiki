<template>
  <span v-if="!isPermanent" class="library-expire-status">
    <a-tooltip v-if="isExpired">
      <template #title>
        <div>{{ t('expired_tip') }}</div>
        <div>
          {{ t('expired_action_prefix') }}
          <a @click="handleClick">{{ t('go_settings') }}</a>
          {{ t('expired_action_suffix') }}
        </div>
      </template>
      <span
        class="status-value expired"
        role="button"
        tabindex="0"
        @click="handleClick"
        @keydown.enter="handleClick"
        @keydown.space.prevent="handleClick"
      >
        <ExclamationCircleFilled class="status-icon" />
        {{ t('expired') }}
      </span>
    </a-tooltip>
    <a-tooltip v-else>
      <template #title>
        <div>{{ t('validity_days', { days: remainingDays }) }}</div>
        <div>{{ t('unavailable_after_expire') }}</div>
      </template>
      <span class="status-value">
        <ClockCircleOutlined class="status-icon" />
        {{ remainingDays }}
      </span>
    </a-tooltip>
  </span>
</template>

<script setup>
import { computed } from 'vue'
import dayjs from 'dayjs'
import { ClockCircleOutlined, ExclamationCircleFilled } from '@ant-design/icons-vue'
import { useI18n } from '@/hooks/web/useI18n'

const { t } = useI18n('components.library-expire-status.index')
const props = defineProps({
  library: {
    type: Object,
    default: () => ({})
  }
})

const expireTime = computed(() => Number(props.library?.expire_time))
const isPermanent = computed(() => {
  return Number(props.library?.is_permanent) !== 0 || !Number.isFinite(expireTime.value) || expireTime.value <= 0
})
const isExpired = computed(() => !isPermanent.value && dayjs().unix() > expireTime.value)
const remainingDays = computed(() => Math.max(0, dayjs.unix(expireTime.value).startOf('day').diff(dayjs().startOf('day'), 'day')))

const handleClick = (event) => {
  event.stopPropagation()
  window.open(`/#/library/details/knowledge-config?id=${props.library.id}`, '_blank', 'noopener')
}
</script>

<style lang="less" scoped>
.library-expire-status {
  display: inline-flex;
  color: #7a8699;
  font-size: 12px;
  line-height: 20px;

  .status-value {
    display: inline-flex;
    align-items: center;
  }

  .status-icon {
    margin-right: 2px;
  }

  .expired {
    color: #fb363f;
    cursor: pointer;
  }
}
</style>
