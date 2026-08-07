<template>
  <div class="skills-page">
    <div class="page-header">
      <div class="page-title">
        <a-segmented :value="activeSkillPage" :options="titleOptios" @change="handleTitleChange" />
      </div>
      <a-dropdown v-if="isAgentPage" placement="bottomRight">
        <a-button type="primary" class="add-btn">
          <template #icon>
            <PlusOutlined />
          </template>
          {{ t('btn_add_skill') }}
          <DownOutlined class="btn-arrow" />
        </a-button>
        <template #overlay>
          <a-menu @click="handleAddMenuClick">
            <!-- <a-menu-item key="skill">新增 Skill</a-menu-item> -->
            <a-menu-item key="tool">{{ t('menu_add_tool') }}</a-menu-item>
            <a-menu-item key="skill">{{ t('menu_add_skill') }}</a-menu-item>
          </a-menu>
        </template>
      </a-dropdown>
      <a-button v-else type="primary" class="add-btn" @click="handleOpenUploadSkillModal">
        <template #icon>
          <PlusOutlined />
        </template>
        上传技能
      </a-button>
    </div>

    <a-row v-if="isAgentPage" :gutter="[16, 16]" class="skill-list agent-skill-list">
      <!-- 固定技能：查询知识库 -->
      <a-col class="skill-col" :xs="24" :md="12" :xl="8" :xxl="6">
        <div class="skill-card">
          <div class="skill-main">
            <div class="skill-title-row">
              <div class="skill-title">{{ t('title_query_knowledge_base') }}</div>
              <div class="skill-tag tool">{{ t('tag_tool') }}</div>
            </div>
            <OverflowTooltip :tooltip-width="320" :title="t('desc_query_knowledge_base')">
              <div class="skill-desc">{{ t('desc_query_knowledge_base') }}</div>
            </OverflowTooltip>
          </div>
          <div class="skill-actions">
            <a-switch
              class="skill-switch"
              size="small"
              :checked="knowledgeEnabled"
              :aria-label="t('title_query_knowledge_base')"
              @change="toggleKnowledge"
            />
          </div>
        </div>
      </a-col>

      <!-- 固定技能：查询本地文档 -->
      <a-col class="skill-col" :xs="24" :md="12" :xl="8" :xxl="6">
        <div class="skill-card">
          <div class="skill-main">
            <div class="skill-title-row">
              <div class="skill-title">{{ t('title_query_local_docs') }}</div>
              <div class="skill-tag skill">{{ t('tag_skill') }}</div>
            </div>
            <OverflowTooltip :tooltip-width="320" :title="t('desc_query_local_docs')">
              <div class="skill-desc">{{ t('desc_query_local_docs') }}</div>
            </OverflowTooltip>
          </div>
          <div class="skill-actions">
            <a-switch
              class="skill-switch"
              size="small"
              :checked="localDocsEnabled"
              :aria-label="t('title_query_local_docs')"
              @change="toggleLocalDocs"
            />
          </div>
        </div>
      </a-col>

      <!-- 固定技能：写文件 -->
      <a-col class="skill-col" :xs="24" :md="12" :xl="8" :xxl="6">
        <div class="skill-card">
          <div class="skill-main">
            <div class="skill-title-row">
              <div class="skill-title">{{ t('title_agent_write_file') }}</div>
              <div class="skill-tag tool">{{ t('tag_tool') }}</div>
            </div>
            <OverflowTooltip :tooltip-width="320" :title="t('desc_agent_write_file')">
              <div class="skill-desc">{{ t('desc_agent_write_file') }}</div>
            </OverflowTooltip>
          </div>
          <div class="skill-actions">
            <a-switch
              class="skill-switch"
              size="small"
              :checked="writeFileEnabled"
              :aria-label="t('title_agent_write_file')"
              @change="toggleWriteFile"
            />
          </div>
        </div>
      </a-col>

      <!-- 固定技能：编辑文件 -->
      <a-col class="skill-col" :xs="24" :md="12" :xl="8" :xxl="6">
        <div class="skill-card">
          <div class="skill-main">
            <div class="skill-title-row">
              <div class="skill-title">{{ t('title_agent_edit_file') }}</div>
              <div class="skill-tag tool">{{ t('tag_tool') }}</div>
            </div>
            <OverflowTooltip :tooltip-width="320" :title="t('desc_agent_edit_file')">
              <div class="skill-desc">{{ t('desc_agent_edit_file') }}</div>
            </OverflowTooltip>
          </div>
          <div class="skill-actions">
            <a-switch
              class="skill-switch"
              size="small"
              :checked="editFileEnabled"
              :aria-label="t('title_agent_edit_file')"
              @change="toggleEditFile"
            />
          </div>
        </div>
      </a-col>

      <!-- 固定技能：执行命令 -->
      <a-col class="skill-col" :xs="24" :md="12" :xl="8" :xxl="6">
        <div class="skill-card">
          <div class="skill-main">
            <div class="skill-title-row">
              <div class="skill-title">{{ t('title_agent_execute') }}</div>
              <div class="skill-tag tool">{{ t('tag_tool') }}</div>
            </div>
            <OverflowTooltip :tooltip-width="320" :title="t('desc_agent_execute')">
              <div class="skill-desc">{{ t('desc_agent_execute') }}</div>
            </OverflowTooltip>
          </div>
          <div class="skill-actions">
            <a-switch
              class="skill-switch"
              size="small"
              :checked="executeEnabled"
              :aria-label="t('title_agent_execute')"
              @change="toggleExecute"
            />
          </div>
        </div>
      </a-col>

      <!-- 固定技能：从商品库推荐商品 -->
      <a-col class="skill-col" :xs="24" :md="12" :xl="8" :xxl="6">
        <div class="skill-card">
          <div class="skill-main">
            <div class="skill-title-row">
              <div class="skill-title">{{ t('title_goods_recommend') }}</div>
              <div class="skill-tag tool">{{ t('tag_tool') }}</div>
            </div>
            <OverflowTooltip :tooltip-width="320" :title="t('desc_goods_recommend')">
              <div class="skill-desc">{{ t('desc_goods_recommend') }}</div>
            </OverflowTooltip>
          </div>
          <div class="skill-actions">
            <span class="recommend-scope-btn" @click="handleOpenScopeModal">{{ t('btn_recommend_scope') }}</span>
            <a-switch
              class="skill-switch"
              size="small"
              :checked="goodsRecommendEnabled"
              :aria-label="t('title_goods_recommend')"
              @change="toggleGoodsRecommend"
            />
          </div>
        </div>
      </a-col>

      <a-col v-if="skillListLoading" :span="24">
        <div class="skill-loading">
          <a-spin />
        </div>
      </a-col>
      <template v-else>
        <!-- 当前 Agent 已绑定的用户技能 -->
        <a-col v-for="item in agentSkills" :key="item.id" class="skill-col" :xs="24" :md="12" :xl="8" :xxl="6">
          <div class="skill-card">
            <div class="skill-main">
              <div class="skill-title-row">
                <div class="skill-title">{{ item.title }}</div>
                <div v-if="item.sourceLabel" class="skill-source-tag" :class="item.sourceClass">
                  {{ item.sourceLabel }}
                </div>
                <div class="skill-tag skill">{{ t('tag_skill') }}</div>
              </div>
              <OverflowTooltip :tooltip-width="320" :title="item.desc">
                <div class="skill-desc">{{ item.desc }}</div>
              </OverflowTooltip>
            </div>
            <div class="skill-actions">
              <a-tooltip :title="t('btn_remove')">
                <button
                  type="button"
                  class="delete-action"
                  :aria-label="t('btn_remove')"
                  @click="handleUnbindSkill(item)"
                >
                  <DeleteOutlined />
                </button>
              </a-tooltip>
            </div>
          </div>
        </a-col>
      </template>

      <!-- 已关联的 WorkFlow 工具列表 -->
      <template v-if="workFlowSkills.length">
        <a-col v-for="item in workFlowSkills" :key="item.id" class="skill-col" :xs="24" :md="12" :xl="8" :xxl="6">
          <div class="skill-card">
            <div class="skill-main">
              <div class="skill-title-row">
                <div class="skill-title">{{ item.name }}</div>
                <div class="skill-tag tool">{{ t('tag_tool') }}</div>
              </div>
              <OverflowTooltip :tooltip-width="320" :title="item.desc || '—'">
                <div class="skill-desc">{{ item.desc || '—' }}</div>
              </OverflowTooltip>
            </div>
            <div class="skill-actions">
              <a-tooltip :title="t('btn_remove')">
                <button
                  type="button"
                  class="delete-action"
                  :aria-label="t('btn_remove')"
                  @click="handleRemoveWorkFlow(item.id)"
                >
                  <DeleteOutlined />
                </button>
              </a-tooltip>
            </div>
          </div>
        </a-col>
      </template>

    </a-row>
    <a-row v-else :gutter="[16, 16]" class="skill-list">
      <a-col v-if="skillListLoading" :span="24">
        <div class="skill-loading">
          <a-spin />
        </div>
      </a-col>
      <a-col v-else-if="librarySkills.length === 0" :span="24">
        <div class="skill-empty">
          <a-empty :description="t('empty_skill')" />
        </div>
      </a-col>
      <template v-else>
        <a-col v-for="item in librarySkills" :key="item.id" class="skill-col" :xs="24" :md="12" :xl="8" :xxl="6">
          <div class="skill-card">
            <div class="skill-main">
              <div class="skill-title-row">
                <div class="skill-title">{{ item.title }}</div>
              </div>
              <OverflowTooltip :tooltip-width="320" :title="item.desc">
                <div class="skill-desc">{{ item.desc }}</div>
              </OverflowTooltip>
            </div>
            <div class="skill-actions">
              <a-tooltip :title="t('btn_edit')">
                <button
                  type="button"
                  class="icon-action"
                  :aria-label="t('btn_edit')"
                  @click="handleEditSkill(item)"
                >
                  <EditOutlined />
                </button>
              </a-tooltip>
              <a-tooltip :title="t('btn_delete')">
                <button
                  type="button"
                  class="delete-action"
                  :aria-label="t('btn_delete')"
                  @click="handleDeleteSkill(item)"
                >
                  <DeleteOutlined />
                </button>
              </a-tooltip>
            </div>
          </div>
        </a-col>
      </template>
    </a-row>
    <SelectSkillModal
      v-model:visible="selectSkillModalVisible"
      :robotId="currentAssistant?.id"
      :refreshKey="selectSkillRefreshKey"
      @create="handleOpenUploadSkillModal"
      @confirm="handleSelectSkillConfirm"
    />

    <!-- 新增 Skill 弹窗 -->
    <AddSkillModal v-model:visible="skillModalVisible" @confirm="handleSkillConfirm" />

    <!-- 新增 Tool 弹窗 -->
    <AddToolModal
      v-model:visible="toolModalVisible"
      :robotId="currentAssistant?.id"
      :workFlowIds="robotInfo?.work_flow_ids"
      @confirm="handleToolConfirm"
    />

    <UploadSkillZipModal
      v-model:visible="uploadSkillZipVisible"
      :robotId="currentAssistant?.id"
      :skill-id="editingSkillId"
      @confirm="handleUploadSkillConfirm"
    />

    <!-- 推荐范围弹窗 -->
    <GoodsRecommendScopeModal v-model:visible="scopeModalVisible" @confirm="handleScopeConfirm" />
  </div>
</template>

<script setup>
import { computed, createVNode, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { DeleteOutlined, DownOutlined, EditOutlined, ExclamationCircleOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { message, Modal } from 'ant-design-vue'
import { useI18n } from '@/hooks/web/useI18n'
import { useClawbotStore } from '@/stores/modules/clawbot'
import { storeToRefs } from 'pinia'
import { getRobotList, relationWorkFlow } from '@/api/robot/index'
import { deleteClawbotSkill, getClawbotSkillList, saveClawbotRobotSkills } from '@/api/clawbot'
import AddSkillModal from './components/AddSkillModal.vue'
import AddToolModal from './components/AddToolModal.vue'
import GoodsRecommendScopeModal from './components/GoodsRecommendScopeModal.vue'
import SelectSkillModal from './components/SelectSkillModal.vue'
import UploadSkillZipModal from './components/UploadSkillZipModal.vue'

const { t } = useI18n('views.clawbot.skills.index')
const route = useRoute()
const router = useRouter()
const clawbotStore = useClawbotStore()
const { robotInfo, currentAssistant } = storeToRefs(clawbotStore)
const { updateClawbotConf, fetchRobotInfo } = clawbotStore

const titleOptios = [
  {
    label: 'Agent技能',
    value: 'agent'
  },
  {
    label: '技能库',
    value: 'library'
  },
  {
    label: '技能生成工具',
    value: 'generator'
  }
]

const activeSkillPage = computed(() => route.query.tab === 'library' ? 'library' : 'agent')
const isAgentPage = computed(() => activeSkillPage.value === 'agent')

const handleTitleChange = (key) => {
  if (key === 'generator') {
    const query = { ...route.query }
    delete query.tab
    router.push({ path: '/clawbot/skill-generate-tool', query })
    return
  }

  router.push({
    path: '/clawbot/skills',
    query: {
      ...route.query,
      tab: key === 'library' ? 'library' : 'agent'
    }
  })
}

// 查询知识库开关：search_knowledge_close=0 表示开启，=1 表示关闭
const knowledgeEnabled = computed(() => !Number(robotInfo.value?.search_knowledge_close || 0))
// 查询本地文档开关
const localDocsEnabled = computed(() => !Number(robotInfo.value?.query_local_docs_close || 0))
// Agent 工具开关：1 表示开启，0 表示关闭
const writeFileEnabled = computed(() => Number(robotInfo.value?.open_agent_write_file_tool || 0) === 1)
const executeEnabled = computed(() => Number(robotInfo.value?.open_agent_execute_tool || 0) === 1)
const editFileEnabled = computed(() => Number(robotInfo.value?.open_agent_edit_file_tool || 0) === 1)

function buildConfPayload(overrides = {}) {
  return {
    id: currentAssistant.value?.id,
    search_knowledge_close: knowledgeEnabled.value ? 0 : 1,
    query_local_docs_close: localDocsEnabled.value ? 0 : 1,
    open_agent_write_file_tool: writeFileEnabled.value ? 1 : 0,
    open_agent_execute_tool: executeEnabled.value ? 1 : 0,
    open_agent_edit_file_tool: editFileEnabled.value ? 1 : 0,
    goods_lib_recommend_switch: goodsRecommendEnabled.value ? 1 : 0,
    goods_lib_recommend_group_ids: robotInfo.value?.goods_lib_recommend_group_ids || '',
    ...overrides
  }
}
// 商品库推荐开关
const goodsRecommendEnabled = computed(() => Number(robotInfo.value?.goods_lib_recommend_switch || 0) === 1)

// 推荐范围弹窗状态
const scopeModalVisible = ref(false)

const toggleKnowledge = async (checked) => {
  if (!currentAssistant.value?.id) {
    return
  }

  await updateClawbotConf(buildConfPayload({ search_knowledge_close: checked ? 0 : 1 }))
}

const toggleLocalDocs = async (checked) => {
  if (!currentAssistant.value?.id) return
  await updateClawbotConf(buildConfPayload({ query_local_docs_close: checked ? 0 : 1 }))
}

const toggleWriteFile = async (checked) => {
  if (!currentAssistant.value?.id) return
  await updateClawbotConf(buildConfPayload({ open_agent_write_file_tool: checked ? 1 : 0 }))
}

const toggleExecute = async (checked) => {
  if (!currentAssistant.value?.id) return
  await updateClawbotConf(buildConfPayload({ open_agent_execute_tool: checked ? 1 : 0 }))
}

const toggleEditFile = async (checked) => {
  if (!currentAssistant.value?.id) return

  await updateClawbotConf(buildConfPayload({ open_agent_edit_file_tool: checked ? 1 : 0 }))
}

const toggleGoodsRecommend = async (checked) => {
  if (!currentAssistant.value?.id) {
    return
  }

  await updateClawbotConf(buildConfPayload({ goods_lib_recommend_switch: checked ? 1 : 0 }))
}

const handleOpenScopeModal = () => {
  scopeModalVisible.value = true
}

const handleScopeConfirm = () => {
  scopeModalVisible.value = false
}

const workFlowSkills = ref([])
const skillListLoading = ref(false)
const allUserSkills = ref([])
const selectSkillModalVisible = ref(false)
const selectSkillRefreshKey = ref(0)
let skillListRequestSeq = 0

// 加载已关联的 WorkFlow 工具列表
const loadWorkFlowSkills = async () => {
  const workFlowIds = robotInfo.value?.work_flow_ids
  if (!workFlowIds) {
    workFlowSkills.value = []
    return
  }
  try {
    const res = await getRobotList({ application_type: 1 })
    if (res && res.res === 0) {
      const allTools = res.data || []
      const savedIds = workFlowIds.split(',').filter(Boolean)
      workFlowSkills.value = allTools
        .filter((item) => savedIds.includes(item.id))
        .map((item) => ({
          id: item.id,
          name: item.robot_name,
          desc: item.robot_intro || ''
        }))
    }
  } catch (err) {
    console.error('加载已关联工具失败', err)
  }
}

// robotInfo 变化时自动加载
watch(
  () => robotInfo.value?.work_flow_ids,
  () => {
    loadWorkFlowSkills()
  },
  { immediate: true }
)

watch(
  () => currentAssistant.value?.id,
  () => {
    loadSkillList()
  },
  { immediate: true }
)

const skillModalVisible = ref(false)
const toolModalVisible = ref(false)
const uploadSkillZipVisible = ref(false)
const editingSkillId = ref(0)

const handleAddMenuClick = ({ key }) => {
  if (key === 'skill') {
    handleOpenSelectSkillModal()
  } else if (key === 'tool') {
    toolModalVisible.value = true
  } else if (key === 'zip') {
    editingSkillId.value = 0
    uploadSkillZipVisible.value = true
  }
}

const handleSkillConfirm = (formData) => {
  // TODO: 对接新增 Skill 接口
  console.log('新增 Skill:', formData)
  loadSkillList()
  if (selectSkillModalVisible.value) {
    selectSkillRefreshKey.value += 1
  }
}

const handleToolConfirm = async () => {
  // 保存成功后刷新 robotInfo 再重新加载列表
  await fetchRobotInfo()
  loadWorkFlowSkills()
}

const handleUploadSkillConfirm = async () => {
  await fetchRobotInfo()
  await loadSkillList()
  if (selectSkillModalVisible.value) {
    selectSkillRefreshKey.value += 1
  }
}

// 移除已关联的 WorkFlow 工具
const handleRemoveWorkFlow = (id) => {
  Modal.confirm({
    title: t('title_remove_tool'),
    icon: createVNode(ExclamationCircleOutlined),
    content: t('msg_confirm_remove_tool'),
    okText: t('btn_confirm'),
    cancelText: t('btn_cancel'),
    okType: 'danger',
    onOk: async () => {
      const currentIds = robotInfo.value?.work_flow_ids?.split(',').filter(Boolean) || []
      const newIds = currentIds.filter((item) => item !== String(id))
      try {
        const res = await relationWorkFlow({
          id: currentAssistant.value?.id,
          work_flow_ids: newIds.join(',')
        })
        if (res && res.res === 0) {
          message.success(t('msg_remove_success'))
          await fetchRobotInfo()
          loadWorkFlowSkills()
        } else {
          message.error(res?.msg || t('msg_remove_failed'))
        }
      } catch (err) {
        console.error('移除工具失败', err)
        message.error(t('msg_remove_failed'))
      }
    }
  })
}

async function loadSkillList() {
  if (!currentAssistant.value?.id) {
    allUserSkills.value = []
    skillListLoading.value = false
    return
  }

  const assistantId = currentAssistant.value.id
  const requestSeq = ++skillListRequestSeq
  skillListLoading.value = true
  allUserSkills.value = []
  try {
    const res = await getClawbotSkillList({ id: assistantId })
    if (requestSeq !== skillListRequestSeq || currentAssistant.value?.id !== assistantId) {
      return
    }
    if (res && res.res === 0) {
      allUserSkills.value = (res.data || []).map((item, index) => ({
        id: `${item.source_type || item.source || 'skill'}-${item.skill_id || 0}-${item.skill_name || index}`,
        skillId: item.skill_id,
        title: item.remark_name || item.skill_name || '—',
        desc: item.intro || item.description || '—',
        selected: Number(item.is_selected) === 1,
        raw: item
      }))
    } else {
      message.error(res?.msg || t('msg_fetch_skill_failed'))
    }
  } catch (err) {
    if (requestSeq !== skillListRequestSeq || currentAssistant.value?.id !== assistantId) {
      return
    }
    console.error('获取 Skill 列表失败', err)
    message.error(err?.msg || t('msg_fetch_skill_failed'))
  } finally {
    if (requestSeq === skillListRequestSeq) {
      skillListLoading.value = false
    }
  }
}

const handleOpenSelectSkillModal = () => {
  selectSkillModalVisible.value = true
}

const handleOpenUploadSkillModal = () => {
  if (!currentAssistant.value?.id) {
    message.error('缺少当前 Agent，请重新从 Agent 页面进入')
    return
  }
  editingSkillId.value = 0
  uploadSkillZipVisible.value = true
}

const handleSelectSkillConfirm = async () => {
  await loadSkillList()
}

const handleEditSkill = (item) => {
  editingSkillId.value = item.skillId
  uploadSkillZipVisible.value = true
}

const handleUnbindSkill = (item) => {
  const targetAssistantId = currentAssistant.value?.id
  if (!targetAssistantId) {
    message.error('缺少当前 Agent，请重新从 Agent 页面进入')
    return
  }

  Modal.confirm({
    title: t('title_remove_skill'),
    icon: createVNode(ExclamationCircleOutlined),
    content: '确认从当前 Agent 移除该技能吗？技能仍会保留在技能库中。',
    okText: t('btn_confirm'),
    cancelText: t('btn_cancel'),
    okType: 'danger',
    onOk: async () => {
      if (String(currentAssistant.value?.id || '') !== String(targetAssistantId)) {
        message.warning('当前 Agent 已切换，请重新执行移除操作')
        return
      }

      try {
        const latestListRes = await getClawbotSkillList({ id: targetAssistantId })
        if (!latestListRes || latestListRes.res !== 0) {
          message.error(latestListRes?.msg || t('msg_fetch_skill_failed'))
          return
        }
        if (String(currentAssistant.value?.id || '') !== String(targetAssistantId)) {
          message.warning('当前 Agent 已切换，请重新执行移除操作')
          return
        }

        const remainingSkillIds = (latestListRes.data || [])
          .filter((skill) => Number(skill.is_selected) === 1 && String(skill.skill_id) !== String(item.skillId))
          .map((skill) => skill.skill_id)

        const res = await saveClawbotRobotSkills({
          id: targetAssistantId,
          skill_ids: remainingSkillIds.join(',')
        })
        if (res && res.res === 0) {
          message.success(t('msg_remove_success'))
          if (String(currentAssistant.value?.id || '') === String(targetAssistantId)) {
            await loadSkillList()
          }
        } else {
          message.error(res?.msg || t('msg_remove_failed'))
        }
      } catch (err) {
        console.error('解除 Agent Skill 绑定失败', err)
        message.error(err?.msg || t('msg_remove_failed'))
      }
    }
  })
}

const handleDeleteSkill = (item) => {
  Modal.confirm({
    title: '删除技能',
    icon: createVNode(ExclamationCircleOutlined),
    content: '删除后将从技能库移除，并清理该技能在全部 Agent 上的绑定，是否继续？',
    okText: t('btn_confirm'),
    cancelText: t('btn_cancel'),
    okType: 'danger',
    onOk: async () => {
      try {
        const res = await deleteClawbotSkill({ skill_id: item.skillId })
        if (res && res.res === 0) {
          message.success('删除成功')
          await loadSkillList()
        } else {
          message.error(res?.msg || '删除失败')
        }
      } catch (err) {
        console.error('删除用户 Skill 失败', err)
        message.error(err?.msg || '删除失败')
      }
    }
  })
}

const agentSkills = computed(() => allUserSkills.value.filter((item) => item.selected))
const librarySkills = computed(() => allUserSkills.value)
</script>

<style lang="less" scoped>
.skills-page {
  --skills-primary: #2475fc;
  --skills-border: #f0f0f0;
  --skills-title: #262626;
  --skills-text: #595959;
  --skills-text-light: #8c8c8c;
  --skills-card-shadow: 0 2px 3px rgba(0, 0, 0, 0.04);

  position: relative;
  min-height: 100vh;
  padding: 22px 24px 32px;
  background: #fff;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    top: -88px;
    right: -46px;
    width: 284px;
    height: 274px;
    border-radius: 999px;
    background: rgba(198, 210, 255, 0.3);
    filter: blur(64px);
    pointer-events: none;
  }
}

.page-header {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;

  .page-title {
    color: var(--skills-title);
    font-size: 20px;
    line-height: 28px;
    .ant-segmented {
      background: #edeff2;
    }
    &::v-deep(.ant-segmented-item-selected) {
      color: #2475fc;
    }
  }

  .add-btn {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    height: 32px;
    padding: 0 16px;
    border: none;
    border-radius: 6px;
    background: var(--skills-primary);
    box-shadow: none;
    font-size: 14px;
    line-height: 22px;

    &:hover,
    &:focus {
      background: #4a8dff;
    }
  }

  .btn-arrow {
    font-size: 12px;
  }
}

.skill-list {
  position: relative;
  z-index: 1;
  align-items: start;
}

.skill-col {
  display: flex;
}

.skill-loading,
.skill-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 120px;
}

.skill-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  grid-template-rows: 24px 44px;
  column-gap: 8px;
  row-gap: 12px;
  min-width: 0;
  width: 100%;
  height: 120px;
  padding: 20px 24px;
  border-radius: 8px;
  border: 1px solid var(--skills-border);
  background: #fff;
  box-shadow: var(--skills-card-shadow);
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;

  &:hover {
    border-color: #659dfc;
    box-shadow: 0 6px 8px rgba(37, 63, 105, 0.16);
  }
}

.skill-main {
  display: contents;
}

.skill-title-row {
  grid-column: 1;
  grid-row: 1;
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 4px;
}

.skill-title {
  overflow: hidden;
  color: var(--skills-title);
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.skill-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 18px;
  padding: 0 6px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 400;
  line-height: 16px;

  &.tool {
    color: #10ae8a;
    background: #cbfaf0;
  }

  &.skill {
    color: var(--skills-primary);
    background: #e5efff;
  }
}

.skill-source-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 18px;
  padding: 0 6px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 400;
  line-height: 16px;

  &.market {
    color: #fa8c16;
    background: #fff1dc;
  }

  &.mine {
    color: #fa8c16;
    background: #fff1dc;
  }
}

.skill-desc {
  grid-column: 1 / -1;
  grid-row: 2;
  display: -webkit-box;
  overflow: hidden;
  height: 44px;
  color: var(--skills-text);
  font-size: 14px;
  line-height: 22px;
  text-overflow: ellipsis;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.agent-skill-list .skill-desc {
  white-space: normal;
}

.skill-actions {
  grid-column: 2;
  grid-row: 1;
  display: flex;
  align-items: center;
  align-self: center;
  flex-shrink: 0;
  gap: 8px;
}

.delete-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  padding: 0;
  border-radius: 6px;
  border: 0;
  background: transparent;
  color: var(--skills-text);
  font: inherit;
  cursor: pointer;
  transition:
    background-color 0.2s ease,
    color 0.2s ease;

  &:hover {
    background: #e4e6eb;
    color: var(--skills-text);
  }

  :deep(svg) {
    width: 16px;
    height: 16px;
  }
}

.icon-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  padding: 0;
  border-radius: 6px;
  border: 0;
  background: transparent;
  color: var(--skills-text);
  font: inherit;
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;

  &:hover {
    background: #e4e6eb;
    color: var(--skills-primary);
  }

}

.skill-text-action {
  height: auto;
  padding: 0;
}

@media (max-width: 768px) {
  .skills-page {
    padding: 20px 16px 24px;
  }

  .page-header {
    flex-direction: column;
    align-items: stretch;
  }

  .page-header .add-btn {
    justify-content: center;
  }

}

.recommend-scope-btn {
  color: var(--skills-primary);
  font-size: 14px;
  cursor: pointer;
  white-space: nowrap;
  user-select: none;
}
</style>
