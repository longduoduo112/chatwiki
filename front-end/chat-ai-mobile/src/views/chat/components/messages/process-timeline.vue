<template>
  <div
    class="process-timeline"
    :class="{ 'is-running': isTimelineRunning, 'is-completed': !isTimelineRunning }"
    v-if="showTimeline"
  >
    <div class="process-timeline-header" @click="toggleTimeline">
      <van-loading
        class="timeline-loading"
        color="#2475fc"
        size="16px"
        type="spinner"
        v-if="isTimelineRunning"
      />
      <van-icon class="timeline-completed-icon" name="passed" v-else />
      <span>{{ timelineTitle }}</span>
      <svg-icon
        name="down-arrow"
        class="timeline-arrow"
        :class="{ expanded: timelineExpanded }"
        v-if="displaySteps.length"
      ></svg-icon>
    </div>

    <div
      class="process-timeline-body"
      v-if="timelineExpanded && displaySteps.length"
    >
      <div
        class="process-step"
        :class="{ 'is-thinking': step.type === 'thinking' }"
        v-for="step in displaySteps"
        :key="step.id"
      >
        <CherryMarkdown
          v-if="step.type === 'thinking'"
          class="process-step-thinking-content"
          :content="getStepText(step)"
        />
        <template v-else>
          <div
            class="process-step-header"
            :class="{ 'is-expandable': canExpandStep(step) }"
            @click="toggleStep(step)"
          >
            <span class="process-step-status">
              <van-loading
                class="status-icon is-running"
                color="#2475fc"
                size="16px"
                type="spinner"
                v-if="step.status === 'running'"
              />
              <van-icon class="is-done" name="passed" v-else />
            </span>

            <div class="process-step-title">
              {{ getStepTitle(step) }}
            </div>
            <svg-icon
              v-if="canExpandStep(step)"
              name="down-arrow"
              class="process-step-arrow"
              :class="{ expanded: isStepExpanded(step) }"
            ></svg-icon>
          </div>

          <div
            class="process-step-detail"
            v-if="isStepExpanded(step) && step.type === 'tool'"
          >
            <div class="process-step-section" v-if="step.type === 'tool' && step.paramsText">
              <span class="process-step-label">{{ t('label_input') }}</span>
              <div class="process-step-scroll">
                <code class="process-step-code">{{ step.paramsText }}</code>
              </div>
            </div>

            <div class="process-step-section" v-if="step.resultText">
              <span class="process-step-label">{{ t('label_output') }}</span>
              <div class="process-step-scroll">
                <CherryMarkdown :content="step.resultText" />
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import CherryMarkdown from '@/components/cherry-markdown/index.vue'
import { useI18n } from '@/hooks/web/useI18n'

const { t } = useI18n('views.chat.components.messages.message-item')

const props = defineProps({
  item: {
    type: Object,
    default: () => ({})
  },
  runningLabel: {
    type: String,
    default: ''
  },
  quoteLoadingVisible: {
    type: Boolean,
    default: false
  }
})

const visibleSteps = computed(() => {
  if (!Array.isArray(props.item?.process_steps)) {
    return []
  }
  return props.item.process_steps.filter((step) => step?.hidden !== true)
})

const displaySteps = computed(() => {
  return visibleSteps.value.filter((step) => {
    if (step?.type === 'tool' || step?.type === 'skill') {
      return !!step.title || !!step.paramsText || !!step.resultText || step.status === 'running'
    }
    if (step?.type === 'thinking') {
      return !!step.contentText || !!step.resultText || step.status === 'running'
    }
    return !!step.resultText || !!step.contentText
  })
})

const hasRunningStep = computed(() => displaySteps.value.some((step) => step?.status === 'running'))
const hasCompletedThinkingStep = computed(() => {
  return displaySteps.value.some((step) => {
    return step?.type === 'thinking' &&
      step?.status !== 'running' &&
      !!(step?.contentText || step?.resultText)
  })
})
const isWaitingForAnswer = computed(() => !!props.item?.startLoading && !props.item?.is_stopped)
const showTimeline = computed(() => {
  return displaySteps.value.length > 0 || (isWaitingForAnswer.value && !props.quoteLoadingVisible)
})
const isTimelineRunning = computed(() => {
  return !props.item?.is_stopped && (
    hasRunningStep.value || (isWaitingForAnswer.value && !hasCompletedThinkingStep.value)
  )
})
const hasStreamingThinkingContent = computed(() => {
  return displaySteps.value.some((step) => {
    return step?.type === 'thinking' &&
      step?.status === 'running' &&
      !!(step?.contentText || step?.resultText)
  })
})
const hasAnswerStartedOrFinished = computed(() => {
  return displaySteps.value.length > 0 && (!props.item?.startLoading || props.item?.is_stopped)
})
const timelineExpanded = ref(false)
const timelineManuallyToggled = ref(false)
const stepExpandedOverrides = ref<Record<string, boolean>>({})
const timelineTitle = computed(() => {
  if (props.item?.is_stopped) {
    return t('label_stopped')
  }
  if (isTimelineRunning.value) {
    return props.runningLabel || t('deep_thinking')
  }
  return t('deep_thinking_completed')
})

const toggleTimeline = () => {
  if (!displaySteps.value.length) {
    return
  }
  timelineExpanded.value = !timelineExpanded.value
  timelineManuallyToggled.value = true
}

watch(hasStreamingThinkingContent, (hasContent) => {
  if (hasContent && !timelineManuallyToggled.value) {
    timelineExpanded.value = true
  }
}, { immediate: true })

watch(hasAnswerStartedOrFinished, (hasFinished, wasFinished) => {
  if (hasFinished && !wasFinished) {
    timelineExpanded.value = false
    timelineManuallyToggled.value = false
  }
})

const isStepExpanded = (step: any) => {
  const stepId = step?.id
  if (stepId && Object.prototype.hasOwnProperty.call(stepExpandedOverrides.value, stepId)) {
    return stepExpandedOverrides.value[stepId]
  }
  return step?.expanded === true
}

const canExpandStep = (step: any) => {
  return step?.type === 'tool' && (!!step.paramsText || !!step.resultText)
}

const toggleStep = (step: any) => {
  if (!step?.id || !canExpandStep(step)) {
    return
  }
  stepExpandedOverrides.value = {
    ...stepExpandedOverrides.value,
    [step.id]: !isStepExpanded(step)
  }
}

const getStepTitle = (step: any) => {
  if (step?.type === 'skill') {
    return t('label_execute_skill', { name: step?.title || '' })
  }
  return t('label_execute_tool', { name: step?.title || step?.eventName || '' })
}

const getStepText = (step: any) => {
  return step?.contentText || step?.resultText || (step?.status === 'running' ? t('msg_thinking') : '')
}
</script>

<style lang="less" scoped>
.process-timeline {
  margin-bottom: 24px;
  color: #8c8c8c;
  font-size: 14px;
  line-height: 22px;
}

.process-timeline.is-running {
  color: #2475fc;

  .process-timeline-header {
    font-size: 14px;
    font-weight: 400;
    line-height: 22px;
  }
}

.process-timeline.is-completed {
  padding: 12px;
  background: #f8f9fc;
  border-radius: 12px;

  .process-timeline-header {
    color: #262626;
    font-size: 16px;
    font-weight: 500;
    line-height: 24px;
  }
}

.process-timeline-header {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  padding: 0;
  color: inherit;
  cursor: pointer;
  user-select: none;
}

.timeline-loading {
  flex: 0 0 auto;
  width: 16px;
  height: 16px;
  transform: translateY(-2px);
}

.timeline-completed-icon {
  flex: 0 0 auto;
  color: #2475fc;
  font-size: 16px;
}

.timeline-arrow,
.process-step-arrow {
  flex: 0 0 auto;
  font-size: 16px;
  color: #8c8c8c;
  transform: rotate(-90deg);
  transition: transform 0.2s;

  &.expanded {
    transform: rotate(0deg);
  }
}

.process-timeline-body {
  padding-top: 12px;
}

.timeline-arrow {
  margin-left: auto;
}

.process-step {
  width: 100%;

  &.is-thinking {
    position: relative;
    padding-left: 20px;

    &::before {
      position: absolute;
      top: 4px;
      bottom: 4px;
      left: 0;
      width: 4px;
      background: #d8dde5;
      border-radius: 8px;
      content: '';
    }
  }
}

.process-step-header {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  min-height: 38px;
  user-select: none;

  &.is-expandable {
    cursor: pointer;
  }
}

.process-step-status {
  flex: 0 0 auto;
  height: 38px;
  display: flex;
  align-items: center;
  justify-content: center;
  .status-icon{
    width: 16px;
    height: 16px;
    font-size: 16px;
  }
  .is-running{
    margin-top: -11px;
  }
  .is-done {
    color: #8c8c8c;
  }
}

.process-step-title {
  flex: 1;
  min-width: 0;
  padding: 8px 0;
  color: #8c8c8c;
  word-break: break-word;
}

.process-step-arrow {
  margin-top: 12px;
}

.process-step-detail {
  padding: 0 0 12px 22px;
}

.process-step-thinking-content,
.process-step-scroll {
  :deep(.cherry-markdown) {
    color: inherit;
    background: transparent;
    margin: 0;
    padding: 0;

    * {
      font-size: 14px;
      line-height: 22px;
    }

    p,
    ul,
    ol,
    blockquote,
    pre,
    table {
      margin-top: 0;
      margin-bottom: 4px;
    }

    > :last-child {
      margin-bottom: 0 !important;
    }
  }
}

.process-step-thinking-content {
  color: #3a4559;

  :deep(.cherry-markdown) {
    word-break: break-word;

    * {
      font-size: 16px;
      line-height: 24px;
    }
  }
}

.process-step-section + .process-step-section {
  margin-top: 12px;
}

.process-step-label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  color: #595959;
}

.process-step-scroll {
  max-height: 200px;
  padding: 8px 12px;
  overflow: auto;
  background: #f5f6f8;
  border-radius: 6px;
  color: #595959;

  &::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }

  &::-webkit-scrollbar-thumb {
    background: #d9d9d9;
    border-radius: 3px;
  }
}

.process-step-code {
  display: block;
  font-size: 12px;
  line-height: 20px;
  white-space: pre-wrap;
  word-break: break-all;
  color: #595959;
}
</style>
