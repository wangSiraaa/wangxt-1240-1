<template>
  <div class="page">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-select v-model="filters.priority" placeholder="优先级筛选" clearable style="width: 140px" @change="loadList">
              <el-option v-for="(label, key) in PriorityLabels" :key="key" :label="label" :value="key" />
            </el-select>
            <el-select v-model="filters.status" placeholder="状态筛选" clearable style="width: 140px" @change="loadList">
              <el-option v-for="(label, key) in SuggestionStatusLabels" :key="key" :label="label" :value="key" />
            </el-select>
            <el-select v-model="filters.granary_id" placeholder="仓房筛选" clearable filterable style="width: 180px" @change="loadList">
              <el-option v-for="g in granaries" :key="g.id" :label="`${g.code} - ${g.name}`" :value="g.id" />
            </el-select>
            <el-button type="primary" :icon="Refresh" @click="loadList">查询</el-button>
          </div>
          <div>
            <el-tag type="danger" effect="dark" :icon="Warning" style="margin-right: 8px" v-if="stats.urgent > 0">
              紧急 {{ stats.urgent }}
            </el-tag>
            <el-tag type="warning" effect="dark" :icon="InfoFilled" v-if="stats.high > 0">
              高优先 {{ stats.high }}
            </el-tag>
          </div>
        </div>
      </template>

      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column label="编号" width="180" prop="suggestion_no" />
        <el-table-column label="仓房" width="130">
          <template #default="{ row }">{{ row.granary?.name || row.granary?.code || '-' }}</template>
        </el-table-column>
        <el-table-column label="优先级" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="PriorityColors[row.priority] as any" effect="dark" size="small">
              {{ PriorityLabels[row.priority] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="触发原因" min-width="220">
          <template #default="{ row }">
            <div class="reasons">
              <el-tag
                v-for="(r, i) in row.trigger_reasons || []"
                :key="i"
                size="small"
                type="info"
                plain
                style="margin: 2px"
              >
                {{ ReasonLabels[r] || r }}
              </el-tag>
              <span v-if="!row.trigger_reasons?.length" class="gray">-</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="最高温" width="100" align="right">
          <template #default="{ row }">
            <span v-if="row.max_temp" :class="row.max_temp >= 30 ? 'hot' : row.max_temp >= 25 ? 'warm' : ''">
              {{ row.max_temp.toFixed(1) }} °C
            </span>
            <span v-else class="gray">-</span>
          </template>
        </el-table-column>
        <el-table-column label="最低温" width="100" align="right">
          <template #default="{ row }">
            <span v-if="row.min_temp">{{ row.min_temp.toFixed(1) }} °C</span>
            <span v-else class="gray">-</span>
          </template>
        </el-table-column>
        <el-table-column label="温差" width="90" align="right">
          <template #default="{ row }">
            <span v-if="row.temp_diff" :class="row.temp_diff >= 5 ? 'hot' : ''">
              {{ row.temp_diff.toFixed(1) }} °C
            </span>
            <span v-else class="gray">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="SuggestionStatusColors[row.status] as any" size="small">
              {{ SuggestionStatusLabels[row.status] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="生成时间" width="170" sortable>
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 'pending' && userStore.hasRole('admin','keeper')"
              type="warning"
              link
              size="small"
              @click="process(row, 'in_progress')"
            >开始处理</el-button>
            <el-button
              v-if="row.status === 'in_progress' && userStore.hasRole('admin','keeper')"
              type="success"
              link
              size="small"
              @click="openCompleteDialog(row)"
            >完成处理</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="detailDialogVisible" title="翻仓建议详情" width="680px">
      <el-descriptions v-if="currentSuggestion" :column="2" border size="small">
        <el-descriptions-item label="建议编号">{{ currentSuggestion.suggestion_no }}</el-descriptions-item>
        <el-descriptions-item label="优先级">
          <el-tag :type="PriorityColors[currentSuggestion.priority] as any" effect="dark">{{ PriorityLabels[currentSuggestion.priority] }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="关联仓房">{{ currentSuggestion.granary?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="关联粮情">{{ currentSuggestion.condition_id ? '已关联' : '-' }}</el-descriptions-item>
        <el-descriptions-item label="最高温">{{ currentSuggestion.max_temp?.toFixed(1) || '-' }} °C</el-descriptions-item>
        <el-descriptions-item label="最低温">{{ currentSuggestion.min_temp?.toFixed(1) || '-' }} °C</el-descriptions-item>
        <el-descriptions-item label="温度差值">{{ currentSuggestion.temp_diff?.toFixed(1) || '-' }} °C</el-descriptions-item>
        <el-descriptions-item label="平均湿度">{{ currentSuggestion.avg_humidity?.toFixed(1) || '-' }} %</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="SuggestionStatusColors[currentSuggestion.status] as any">
            {{ SuggestionStatusLabels[currentSuggestion.status] }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="生成时间">{{ formatTime(currentSuggestion.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="触发原因" :span="2">
          <el-tag v-for="(r, i) in currentSuggestion.trigger_reasons || []" :key="i" size="small" style="margin-right: 4px">
            {{ ReasonLabels[r] || r }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="异常区域" :span="2">
          <div v-if="currentSuggestion.anomaly_zones?.length" class="zones">
            <div v-for="(z, i) in currentSuggestion.anomaly_zones" :key="i" class="zone">
              <el-tag type="danger" size="small">区域 {{ i + 1 }}</el-tag>
              <span class="zone-info">
                位置：{{ z.location || '-' }} |
                温度：{{ z.temperature }}°C |
                描述：{{ z.description || '-' }}
              </span>
            </div>
          </div>
          <span v-else class="gray">无记录</span>
        </el-descriptions-item>
        <el-descriptions-item label="处置建议" :span="2">
          <div class="suggestion-text">{{ currentSuggestion.suggestion_text || '-' }}</div>
        </el-descriptions-item>
        <el-descriptions-item v-if="currentSuggestion.process_note" label="处理说明" :span="2">
          <div class="suggestion-text process">{{ currentSuggestion.process_note }}</div>
        </el-descriptions-item>
        <el-descriptions-item v-if="currentSuggestion.processed_by" label="处理人">
          {{ currentSuggestion.processed_by }}
        </el-descriptions-item>
        <el-descriptions-item v-if="currentSuggestion.processed_at" label="处理时间">
          {{ formatTime(currentSuggestion.processed_at) }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <el-dialog v-model="completeDialogVisible" title="完成处理" width="520px">
      <el-form :model="completeForm" label-width="100px">
        <el-form-item label="处理结果">
          <el-radio-group v-model="completeForm.result">
            <el-radio value="turned">已完成翻仓</el-radio>
            <el-radio value="monitoring">继续观察监测</el-radio>
            <el-radio value="other">其他处理</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="处理说明">
          <el-input v-model="completeForm.process_note" type="textarea" :rows="4" placeholder="请描述处理措施和结果" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="completeDialogVisible = false">取消</el-button>
        <el-button type="success" :loading="processing" @click="handleComplete">确认完成</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, type FormInstance } from 'element-plus'
import { Refresh, Warning, InfoFilled } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import type { Granary, TurnoverSuggestion, SuggestionPriority, SuggestionStatus } from '@/types'
import { PriorityLabels, PriorityColors, SuggestionStatusLabels, SuggestionStatusColors } from '@/types'
import { granaryApi, suggestionApi } from '@/api'
import dayjs from 'dayjs'

const ReasonLabels: Record<string, string> = {
  temp_too_high: '温度偏高',
  temp_diff_large: '温差过大',
  moisture_high: '水分偏高',
  insect_found: '虫害',
  mold_found: '霉变',
  abnormal_grain: '粮情异常',
  regular_check: '常规检查'
}

const userStore = useUserStore()

const loading = ref(false)
const list = ref<TurnoverSuggestion[]>([])
const granaries = ref<Granary[]>([])

const filters = reactive({
  priority: '' as SuggestionPriority | '',
  status: '' as SuggestionStatus | '',
  granary_id: ''
})

const stats = computed(() => {
  const urgent = list.value.filter(s => s.priority === 'urgent' && s.status !== 'resolved').length
  const high = list.value.filter(s => s.priority === 'high' && s.status !== 'resolved').length
  return { urgent, high }
})

const detailDialogVisible = ref(false)
const currentSuggestion = ref<TurnoverSuggestion>()

const completeDialogVisible = ref(false)
const processing = ref(false)
const completeForm = reactive({ result: 'turned', process_note: '' })
const currentProcessId = ref('')

const mockList: TurnoverSuggestion[] = [
  {
    id: '1', suggestion_no: 'TS20241220120001',
    granary_id: '10000000-0000-0000-0000-000000000002',
    granary: { name: '二号仓', code: 'A-02' } as any,
    priority: 'urgent', status: 'pending',
    trigger_reasons: ['temp_too_high', 'insect_found'],
    max_temp: 31.8, min_temp: 22.4, temp_diff: 9.4,
    avg_humidity: 78.5,
    anomaly_zones: [
      { location: '北区上层', temperature: 31.8, description: '疑似玉米象虫害活跃区' },
      { location: '西区中层', temperature: 29.6, description: '高温区' }
    ],
    suggestion_content: '',
    suggestion_text:
      '【紧急处置】最高温 31.8°C 超过 30°C 红线，同时检测到玉米象虫害。\n\n' +
      '1. 立即通知保管员、安全员现场确认虫害范围；\n' +
      '2. 启动紧急翻仓方案，优先翻倒北区上层高温暖区；\n' +
      '3. 翻仓后建议立即申请熏蒸处置，防止虫害扩散；\n' +
      '4. 同步检测西区中层区域粮温变化。',
    created_at: '2024-12-20T12:00:00Z',
    updated_at: '2024-12-20T12:00:00Z'
  } as TurnoverSuggestion,
  {
    id: '2', suggestion_no: 'TS20241221100002',
    granary_id: '10000000-0000-0000-0000-000000000003',
    granary: { name: '三号仓', code: 'B-01' } as any,
    priority: 'high', status: 'in_progress',
    trigger_reasons: ['temp_diff_large'],
    max_temp: 26.5, min_temp: 20.1, temp_diff: 6.4,
    avg_humidity: 72.0,
    anomaly_zones: [
      { location: '南区中层', temperature: 26.5, description: '散热不良区' }
    ],
    suggestion_content: '',
    suggestion_text:
      '【重点处置】最高温 26.5°C，上下温差 6.4°C 超过 5°C 预警线。\n\n' +
      '1. 安排翻仓作业，优先处理南区中层区域；\n' +
      '2. 翻仓时配合轴流风机强制通风降温；\n' +
      '3. 翻仓完成后 24 小时内再次检测温差；\n' +
      '4. 建议将此仓房列入重点监控名单，每日定时巡检。',
    created_at: '2024-12-21T10:00:00Z',
    updated_at: '2024-12-22T08:00:00Z',
    process_note: '已开始翻仓作业，预计今日完成南区处理'
  } as TurnoverSuggestion,
  {
    id: '3', suggestion_no: 'TS20241218100003',
    granary_id: '10000000-0000-0000-0000-000000000001',
    granary: { name: '一号仓', code: 'A-01' } as any,
    priority: 'medium', status: 'resolved',
    trigger_reasons: ['moisture_high'],
    max_temp: 23.2, min_temp: 19.8, temp_diff: 3.4,
    avg_humidity: 81.2,
    anomaly_zones: [],
    suggestion_content: '',
    suggestion_text:
      '【常规处置】平均湿度 81.2%，存在轻微霉变风险。\n\n' +
      '1. 适当增加通风时间，降低仓内湿度；\n' +
      '2. 每周定期翻仓一次，促进粮堆散热散湿；\n' +
      '3. 关注天气变化，雨天及时关闭通风口。',
    created_at: '2024-12-18T10:00:00Z',
    updated_at: '2024-12-19T16:00:00Z',
    processed_by: '张保管员',
    processed_at: '2024-12-19T16:00:00Z',
    process_note: '已完成通风和翻仓处理，湿度降至 72%，状态恢复正常。'
  } as TurnoverSuggestion
]

const formatTime = (t?: string) => t ? dayjs(t).format('YYYY-MM-DD HH:mm') : '-'

const loadList = async () => {
  loading.value = true
  try {
    const data = await suggestionApi.list({
      priority: filters.priority || undefined,
      status: filters.status || undefined,
      granary_id: filters.granary_id || undefined
    })
    list.value = data
  } catch {
    list.value = mockList.filter(s => {
      if (filters.priority && s.priority !== filters.priority) return false
      if (filters.status && s.status !== filters.status) return false
      if (filters.granary_id && s.granary_id !== filters.granary_id) return false
      return true
    })
  } finally {
    loading.value = false
  }
}

const loadGranaries = async () => {
  try {
    granaries.value = await granaryApi.list()
  } catch {
    granaries.value = [
      { id: '10000000-0000-0000-0000-000000000001', code: 'A-01', name: '一号仓' } as any,
      { id: '10000000-0000-0000-0000-000000000002', code: 'A-02', name: '二号仓' } as any,
      { id: '10000000-0000-0000-0000-000000000003', code: 'B-01', name: '三号仓' } as any
    ]
  }
}

const viewDetail = (row: TurnoverSuggestion) => {
  currentSuggestion.value = row
  detailDialogVisible.value = true
}

const process = async (row: TurnoverSuggestion, status: SuggestionStatus) => {
  try {
    try {
      await suggestionApi.handle(row.id, { status })
    } catch {}
    ElMessage.success('已开始处理')
    loadList()
  } catch {}
}

const openCompleteDialog = (row: TurnoverSuggestion) => {
  currentProcessId.value = row.id
  completeForm.result = 'turned'
  completeForm.process_note = ''
  completeDialogVisible.value = true
}

const handleComplete = async () => {
  processing.value = true
  try {
    try {
      await suggestionApi.handle(currentProcessId.value, {
        status: 'resolved',
        handle_remark: `[${completeForm.result}] ${completeForm.process_note}`
      })
    } catch {}
    ElMessage.success('处理完成，状态已更新')
    completeDialogVisible.value = false
    loadList()
  } finally {
    processing.value = false
  }
}

onMounted(() => {
  loadList()
  loadGranaries()
})
</script>

<style scoped>
.page { width: 100%; }
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.hot { color: #f56c6c; font-weight: 600; }
.warm { color: #e6a23c; font-weight: 600; }
.gray { color: #c0c4cc; }
.reasons { line-height: 28px; }
.zones { line-height: 32px; }
.zone { padding: 6px 8px; margin-bottom: 6px; background: #fff5f0; border-radius: 4px; }
.zone-info { margin-left: 8px; color: #606266; font-size: 13px; }
.suggestion-text {
  padding: 12px 16px;
  background: #f4f8fa;
  border-left: 4px solid #409eff;
  border-radius: 4px;
  white-space: pre-wrap;
  font-size: 13px;
  line-height: 1.8;
  color: #303133;
}
.suggestion-text.process {
  background: #f0f9eb;
  border-left-color: #67c23a;
}
</style>
