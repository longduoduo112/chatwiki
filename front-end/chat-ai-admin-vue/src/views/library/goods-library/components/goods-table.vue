<template>
  <div ref="tableRef" class="goods-table">
    <a-table
      :columns="columns"
      :data-source="rows"
      :loading="loading"
      :pagination="tablePagination"
      @change="handleTableChange"
      :row-key="(record) => record.id"
      :scroll="{ x: 1600 }"
      :custom-row="getCustomRow"
      :row-class-name="getRowClassName"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'goods'">
          <div
            class="cell-inner goods-inner"
            :class="isActiveCell(record.id, 'basic_info') ? 'cell-active' : ''"
            @click.stop="emit('edit-basic-info', record)"
          >
            <div class="goods-image-box">
              <img v-if="getImageUrl(record)" :src="getImageUrl(record)" alt="" />
              <div v-else class="image-empty">{{ t('table.no_image') }}</div>
            </div>
            <div class="goods-summary">
              <div class="goods-name" :title="record.goods_name || record.name">
                {{ record.goods_name || record.name || '-' }}
              </div>
              <div class="goods-price">{{ formatPrice(record.price) }}</div>
            </div>
          </div>
        </template>

        <template v-else-if="column.key === 'description'">
          <div
            class="cell-inner description-inner"
            @dblclick.stop
          >
            <div class="description-content">
              <a-tooltip
                v-if="record.description"
                :mouse-enter-delay="0.3"
                :overlay-style="{ maxWidth: '400px' }"
                overlay-class-name="goods-cell-tooltip"
              >
                <template #title>
                  <div class="tooltip-pre-line">{{ record.description }}</div>
                </template>
                <div class="line-clamp description-text">
                  {{ record.description }}
                </div>
              </a-tooltip>
              <div v-else class="line-clamp description-text placeholder">
                {{ t('table.description_empty') }}
              </div>

              <div v-if="getDescriptionImages(record).length" class="description-images">
                <img
                  v-for="(image, index) in getDescriptionImages(record)"
                  :key="`${record.id}-description-image-${index}`"
                  :src="image"
                  alt=""
                  @click.stop="handlePreviewDescriptionImages(record, image)"
                />
              </div>
            </div>
          </div>
        </template>

        <template v-else-if="column.key === 'goods_card'">
          <div class="cell-inner goods-card-inner">
            <a-button type="link" class="goods-card-setting" @click.stop="emit('edit-card', record)">
              {{ t('goods_card.setting') }}
            </a-button>
          </div>
        </template>

        <template v-else-if="column.key === 'qa'">
          <div
            class="cell-inner text-inner"
            :class="isActiveCell(record.id, 'qa') ? 'cell-active' : ''"
            @click.stop="emit('edit-field', { row: record, fieldKey: 'qa', fieldLabel: t('table.qa'), mode: 'textarea' })"
          >
            <a-tooltip
              v-if="record.qa"
              :mouse-enter-delay="0.3"
              :overlay-style="{ maxWidth: '400px' }"
              overlay-class-name="goods-cell-tooltip"
            >
              <template #title>
                <div class="tooltip-pre-line">{{ record.qa }}</div>
              </template>
              <div class="line-clamp">
                {{ record.qa }}
              </div>
            </a-tooltip>
            <div v-else class="line-clamp placeholder">
              {{ t('table.qa_placeholder') }}
            </div>
          </div>
        </template>

        <template v-else-if="column.key === 'custom_info'">
          <div
            class="cell-inner text-inner"
            :class="isActiveCell(record.id, 'custom_info') ? 'cell-active' : ''"
            @click.stop="emit('edit-field', { row: record, fieldKey: 'custom_info', fieldLabel: t('custom_info.title'), mode: 'custom_info' })"
          >
            <a-tooltip
              v-if="record.custom_info"
              :mouse-enter-delay="0.3"
              :overlay-style="{ maxWidth: '400px' }"
              overlay-class-name="goods-cell-tooltip"
            >
              <template #title>
                <div class="tooltip-pre-line">{{ record.custom_info }}</div>
              </template>
              <div class="line-clamp">
                {{ record.custom_info }}
              </div>
            </a-tooltip>
            <div v-else class="line-clamp placeholder">
              {{ t('custom_info.empty') }}
            </div>
          </div>
        </template>

        <template v-else-if="column.key === 'actions'">
          <div class="actions-cell" @click.stop>
            <a-switch
              :checked="record.switch_status === 1"
              :checked-children="t('table.enabled')"
              :un-checked-children="t('table.disabled')"
              @change="(checked) => emit('toggle-status', { row: record, checked })"
            />
            <a-button type="link" class="edit-btn" @click.stop="emit('edit-row', record)">
              {{ t('table.edit') }}
            </a-button>
            <a-button type="link" class="delete-btn" @click.stop="emit('delete-row', record)">
              {{ t('confirm.delete_btn') }}
            </a-button>
          </div>
        </template>
      </template>
    </a-table>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { api as viewerApi } from 'v-viewer'
import { useI18n } from '@/hooks/web/useI18n'

const { t } = useI18n('views.library.goods-library.index')

const props = defineProps({
  rows: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  },
  selectedRowId: {
    type: [String, Number],
    default: ''
  },
  activeCell: {
    type: Object,
    default: () => ({})
  },
  pagination: {
    type: Object,
    default: () => ({ page: 1, size: 20, total: 0 })
  }
})

const emit = defineEmits(['hover-cell', 'edit-field', 'toggle-status', 'delete-row', 'select-row', 'edit-row', 'edit-basic-info', 'edit-card', 'change'])

const tableRef = ref(null)
const scrollY = ref(400)
let resizeObserver = null

const HEADER_HEIGHT = 54
const PAGINATION_HEIGHT = 56
const SAFE_GAP = 8

onMounted(() => {
  const container = tableRef.value?.closest('.main-content')
  if (!container) return

  resizeObserver = new ResizeObserver((entries) => {
    for (const entry of entries) {
      const height = entry.contentRect.height
      scrollY.value = Math.max(height - HEADER_HEIGHT - PAGINATION_HEIGHT - SAFE_GAP, 200)
    }
  })
  resizeObserver.observe(container)
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
})

const tablePagination = computed(() => ({
  current: props.pagination.page,
  pageSize: props.pagination.size,
  total: props.pagination.total,
  showSizeChanger: true,
  showQuickJumper: true,
  pageSizeOptions: ['10', '20', '50'],
  size: 'default'
}))

const handleTableChange = (pagination) => {
  emit('change', {
    page: pagination.current,
    size: pagination.pageSize
  })
}

const columns = computed(() => [
  {
    title: t('table.goods'),
    key: 'goods',
    width: 280
  },
  {
    title: t('table.description'),
    key: 'description',
    width: 360
  },
  {
    title: t('table.qa'),
    key: 'qa',
    width: 263
  },
  {
    title: t('goods_card.table_title'),
    key: 'goods_card',
    width: 160
  },
  {
    title: t('custom_info.title'),
    key: 'custom_info',
    width: 383
  },
  {
    title: t('table.actions'),
    key: 'actions',
    width: 156,
    fixed: 'right',
    className: 'actions-fixed-column'
  }
])


const normalizeId = (value) => {
  if (value === undefined || value === null || value === '') {
    return ''
  }

  return String(value)
}

const getCustomRow = (record) => ({
  onClick: () => {
    emit('select-row', record)
  },
  onDblclick: () => {
    emit('edit-row', record)
  }
})

const getRowClassName = (record) => {
  return normalizeId(record.id) === normalizeId(props.selectedRowId) ? 'row-selected' : ''
}

const isActiveCell = (rowId, field) => {
  return props.activeCell?.rowId === rowId && props.activeCell?.field === field
}

const getImageList = (record) => {
  if (Array.isArray(record.images) && record.images.length) {
    return record.images
  }

  return []
}

const getImageUrl = (record) => {
  const images = getImageList(record)
  return images[0] || ''
}

const getDescriptionImages = (record) => {
  return getImageList(record)
}

const handlePreviewDescriptionImages = (record, image) => {
  const images = getDescriptionImages(record)
  const initialViewIndex = images.indexOf(image)

  viewerApi({
    images,
    options: {
      initialViewIndex: initialViewIndex >= 0 ? initialViewIndex : 0,
      toolbar: true,
      title: false,
      movable: true,
      zoomable: true,
      rotatable: true,
      scalable: true
    }
  })
}

const formatPrice = (price) => {
  if (price === undefined || price === null || price === '') return '-'
  return `¥ ${price}`
}
</script>

<style lang="less" scoped>
.goods-table {
  width: 100%;
  height: 100%;
  overflow: hidden;
  overflow-y: auto;

  :deep(.ant-table-cell) {
    padding: 4px;
    background: #fff;
    border-bottom: 1px solid #e8e8e8;
  }

  :deep(.ant-table-thead > tr > th) {
    height: 54px;
    padding: 16px;
    background: #f5f5f5;
    color: #262626;
    font-size: 14px;
    line-height: 20px;
    font-weight: 400;
  }

  :deep(.ant-table-tbody > tr) {
    cursor: pointer;
  }

  :deep(.ant-table-tbody > tr > td) {
    padding: 4px;
  }

  :deep(.ant-table-tbody > tr:hover > td) {
    background: #fff !important;
  }

  :deep(.ant-table-placeholder .ant-table-cell) {
    padding: 80px 0;
    border-bottom: none;
  }

  :deep(.ant-table-cell-fix-right),
  :deep(.ant-table-cell-fix-right-first) {
    z-index: 3;
    background: #fff;
  }

  :deep(.ant-table-thead > tr > .ant-table-cell-fix-right),
  :deep(.ant-table-thead > tr > .ant-table-cell-fix-right-first) {
    background: #f5f5f5;
  }

  :deep(.ant-table-thead > tr > .actions-fixed-column),
  :deep(.ant-table-tbody > tr > .actions-fixed-column) {
    box-shadow: -8px 0 12px rgba(0, 0, 0, 0.06);
  }

  :deep(.ant-table-pagination.ant-pagination) {
    margin: 12px 0 0;
    padding: 0 4px;
  }

  .cell-inner {
    position: relative;
    width: 100%;
    height: 100%;
    padding: 12px;
    border-radius: 6px;
    background: #fff;
    transition: background-color 0.2s;

    &:hover {
      background: #e5efff;
    }
  }

  .goods-inner {
    display: flex;
    align-items: center;
    gap: 12px;
    height: 116px;
    overflow: hidden;
  }

  .goods-image-box {
    flex-shrink: 0;
    width: 56px;
    height: 56px;
    overflow: hidden;
    border-radius: 8px;
    background: #fafafa;

    img {
      display: block;
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  .image-empty {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    color: #bfbfbf;
    font-size: 12px;
  }

  .goods-summary {
    min-width: 0;
  }

  .goods-name {
    overflow: hidden;
    color: #262626;
    font-size: 14px;
    font-weight: 500;
    line-height: 22px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .goods-price {
    margin-top: 4px;
    color: #f53f3f;
    font-size: 14px;
    line-height: 20px;
  }

  .description-inner {
    height: 116px;
    overflow: hidden;

    &:hover {
      background: #fff;
    }
  }

  .description-content {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .description-content .description-text {
    -webkit-line-clamp: 2;
    line-clamp: 2;
  }

  .description-images {
    display: flex;
    gap: 6px;

    img {
      width: 32px;
      height: 32px;
      border-radius: 6px;
      cursor: pointer;
      object-fit: cover;
    }
  }

  .goods-card-inner {
    display: flex;
    align-items: center;
    height: 116px;
  }

  .goods-card-setting {
    padding: 0;
  }

  .text-inner {
    height: 116px;
    overflow: hidden;
  }

  .line-clamp {
    color: #595959;
    font-size: 14px;
    line-height: 20px;
    word-break: break-word;
    display: -webkit-box;
    -webkit-line-clamp: 4;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;

    &.placeholder {
      color: #bfbfbf;
    }
  }

  .actions-cell {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    min-height: 122px;
    padding: 12px;
  }

  .delete-btn {
    padding: 0;
    color: #f53f3f;
  }

  .edit-btn {
    padding: 0;
  }
}
</style>

<style lang="less">
.goods-cell-tooltip {
  .tooltip-pre-line {
    max-width: 400px;
    max-height: 300px;
    overflow-y: auto;
    word-break: break-word;
    white-space: pre-line;

    &::-webkit-scrollbar {
      width: 5px;
    }

    &::-webkit-scrollbar-track {
      background: transparent;
    }

    &::-webkit-scrollbar-thumb {
      background: rgba(255, 255, 255, 0.3);
      border-radius: 3px;

      &:hover {
        background: rgba(255, 255, 255, 0.5);
      }
    }
  }
}
</style>
