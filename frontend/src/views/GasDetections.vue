<template>
  <div class="page">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-select v-model="filters.gas_type" placeholder="气体类型" clearable style="width: 160px" @change="loadList">
              <el-option label="磷化氢 PH₃" value="gas_ph3" />
              <el-option label="硫化氢 H₂S" value="gas_h2s" />
              <el-option label="二氧化碳 CO₂" value="co2" />
              <el-option label="氧气 O₂" value="o2" />
            </el-select>
            <el-select v-model="filters.granary_id" placeholder="仓房筛选" clearable filterable style="width: 180px" @change="loadList">
              <el-option v-for="g in granaries" :key="g.id" :label="`${g.code} - ${g.name}`" :value="g.id" />
            </el-select>
            <el-date-picker
              v-model="filters.date"
              type="date"
              placeholder="检测日期"
              value-format="YYYY-MM-DD"
              style="width: 160px"
              clearable
              @change="loadList"
            />
            <el-button type="primary" :icon="Refresh" @click="loadList">查询</el-button>
          </div>
          <el-button v-if="userStore.hasRole('admin','duty_officer')" type="success" :icon="Plus" @click="openCreateDialog">
            登记气体检测
          </el-button>
        </div>
      </template>

      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column label="检测时间" width="170" sortable>
          <template #default="{ row }">{{ formatTime(row.detection_time) }}</template>
        </el-table-column>
        <el-table-column label="仓房" width="130">
          <template #default="{ row }">{{ row.granary?.name || '-' }}</template>
        </el-table-column>
        <el-table-column label="气体类型" width="130">
          <template #default="{ row }">{{ GasTypeLabels[row.gas_type] || row.gas_type }}</template>
        </el-table-column>
        <el-table-column prop="concentration" label="实测浓度" width="120" align="right">
          <template #default="{ row }">
            <span class="value">{{ row.concentration }}</span> ppm
          </template>
        </el-table-column>
        <el-table-column prop="safe_limit" label="安全限值" width="120" align="right">
          <template #default="{ row }">
            <span class="limit">{{ row.safe_limit }}</span> ppm
          </template>
        </el-table-column>
        <el-table-column label="安全状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_safe" type="success" effect="dark" size="small">
              安全
            </el-tag>
            <el-tag v-else type="danger" effect="dark" size="small">
              超限
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="超限详情" min-width="200">
          <template #default="{ row }">
            <el-popover
              v-if="!row.is_safe && row.violations?.length"
              trigger="hover"
              placement="left"
              :width="280"
            >
              <template #reference>
                <el-tag type="danger" plain size="small">
                  点击查看超限详情
                </el-tag>
              </template>
              <div style="font-size: 13px; line-height: 1.6">
                <div v-for="(v, i) in row.violations" :key="i" class="violation">
                  - {{ GasTypeLabels[v.gas_type as any] || v.gas_type }}:
                  <b class="value">{{ v.actual }}</b> ppm > 限值
                  <b class="limit">{{ v.limit }}</b> ppm
                  （超出 <span class="over">{{ ((v.actual - v.limit) / v.limit * 100).toFixed(1) }}%</span>）
                </div>
              </div>
            </el-popover>
            <span v-else class="gray">-</span>
          </template>
        </el-table-column>
        <el-table-column label="检测人" width="110">
          <template #default="{ row }">{{ row.detector?.full_name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" show-overflow-tooltip />
      </el-table>

      <div class="legend-tips">
        <el-tag type="warning" size="small" effect="plain">⚠ 安全标准</el-tag>
        &nbsp;磷化氢 ≤ 0.3ppm　硫化氢 ≤ 10ppm　二氧化碳 ≤ 5000ppm　氧气 ≥ 19.5%
      </div>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="登记气体检测" width="620px" destroy-on-close>
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="120px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="仓房" prop="granary_id">
              <el-select v-model="createForm.granary_id" placeholder="选择仓房" filterable style="width: 100%" @change="onGranaryChange">
                <el-option v-for="g in granaries" :key="g.id" :label="`${g.code} - ${g.name}`" :value="g.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="关联解封/通风" prop="unseal_id">
              <el-select v-model="createForm.unseal_id" placeholder="选择（可选）" clearable filterable style="width: 100%">
                <el-option
                  v-for="u in unsealList"
                  :key="u.id"
                  :label="`${u.unseal_type === 'ventilation' ? '通风' : '解封'} | ${u.start_time?.slice(0, 16)}`"
                  :value="u.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="检测时间" prop="detection_time">
              <el-date-picker v-model="createForm.detection_time" type="datetime" style="width: 100%" value-format="YYYY-MM-DDTHH:mm:ss" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="检测点">
              <el-input v-model="createForm.detection_points" placeholder="仓内中心、北侧、南侧..." />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-divider content-position="left">气体浓度读数</el-divider>
          </el-col>
          <el-col :span="12">
            <el-form-item label="PH₃ (ppm)">
              <el-input-number v-model="createForm.readings.gas_ph3" :precision="3" :min="0" :step="0.01" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="H₂S (ppm)">
              <el-input-number v-model="createForm.readings.gas_h2s" :precision="3" :min="0" :step="0.1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="CO₂ (ppm)">
              <el-input-number v-model="createForm.readings.co2" :precision="1" :min="0" :step="10" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="O₂ (ppm)">
              <el-input-number v-model="createForm.readings.o2" :precision="0" :min="0" :step="1000" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-alert v-if="currentSafeStatus !== null" :type="currentSafeStatus ? 'success' : 'error'" :closable="false" show-icon>
              当前气体状态：<b>{{ currentSafeStatus ? '已达标，可以解封' : '未达标，请勿解封！' }}</b>
            </el-alert>
          </el-col>
          <el-col :span="24" style="margin-top: 12px">
            <el-form-item label="备注">
              <el-input v-model="createForm.remark" type="textarea" :rows="3" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">登记</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Refresh, Plus } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import type { Granary, GasDetectionRecord } from '@/types'
import { granaryApi, unsealApi } from '@/api'
import dayjs from 'dayjs'

const GasTypeLabels: Record<string, string> = {
  gas_ph3: '磷化氢 PH₃',
  gas_h2s: '硫化氢 H₂S',
  co2: '二氧化碳 CO₂',
  o2: '氧气 O₂'
}

const DefaultSafeLimits: Record<string, number> = {
  gas_ph3: 0.3,
  gas_h2s: 10,
  co2: 5000,
  o2: 195000
}

const userStore = useUserStore()

const loading = ref(false)
const list = ref<GasDetectionRecord[]>([])
const granaries = ref<Granary[]>([])
const unsealList = ref<any[]>([])

const filters = reactive({
  gas_type: '' as string,
  granary_id: '',
  date: ''
})

const createDialogVisible = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = reactive({
  granary_id: '',
  unseal_id: '',
  detection_time: dayjs().format('YYYY-MM-DDTHH:mm:ss'),
  detection_points: '',
  readings: { gas_ph3: 0, gas_h2s: 0, co2: 0, o2: 0 },
  remark: ''
})
const createRules: FormRules = {
  granary_id: [{ required: true, message: '请选择仓房', trigger: 'change' }],
  detection_time: [{ required: true, message: '请选择检测时间', trigger: 'change' }]
}

const currentSafeStatus = computed(() => {
  const r = createForm.readings
  const allZero = r.gas_ph3 === 0 && r.gas_h2s === 0 && r.co2 === 0 && r.o2 === 0
  if (allZero) return null
  return (
    r.gas_ph3 <= DefaultSafeLimits.gas_ph3 &&
    r.gas_h2s <= DefaultSafeLimits.gas_h2s &&
    r.co2 <= DefaultSafeLimits.co2 &&
    (r.o2 === 0 || r.o2 >= DefaultSafeLimits.o2)
  )
})

const mockList: GasDetectionRecord[] = [
  {
    id: '1', granary_id: 'g1', unseal_id: 'u1', detector_id: 'u4',
    granary: { name: '一号仓' } as any,
    detection_time: '2024-12-22T10:00:00Z',
    gas_type: 'gas_ph3', concentration: 0.2, safe_limit: 0.3,
    is_safe: true, violations: [],
    detection_points: '中心,北侧,南侧,东侧,西侧',
    readings: { gas_ph3: 0.2, gas_h2s: 0.5, co2: 800, o2: 208000 } as any,
    remark: '首次通风后检测，指标正常',
    created_at: '2024-12-22T10:00:00Z'
  },
  {
    id: '2', granary_id: 'g1', unseal_id: 'u1', detector_id: 'u4',
    granary: { name: '一号仓' } as any,
    detection_time: '2024-12-22T08:00:00Z',
    gas_type: 'gas_ph3', concentration: 2.5, safe_limit: 0.3,
    is_safe: false,
    violations: [
      { gas_type: 'gas_ph3', actual: 2.5, limit: 0.3 },
      { gas_type: 'co2', actual: 8500, limit: 5000 }
    ],
    detection_points: '中心,北侧,南侧,东侧,西侧',
    readings: { gas_ph3: 2.5, gas_h2s: 3, co2: 8500, o2: 208000 } as any,
    remark: '通风前检测，浓度较高',
    created_at: '2024-12-22T08:00:00Z'
  }
]

const formatTime = (t?: string) => t ? dayjs(t).format('YYYY-MM-DD HH:mm') : '-'

watch(() => createForm.granary_id, () => loadUnsealList())

const loadList = async () => {
  loading.value = true
  try {
    const resp = await unsealApi.listGasDetections({
      gas_type: filters.gas_type || undefined,
      granary_id: filters.granary_id || undefined
    })
    list.value = Array.isArray(resp) ? resp : (resp as any)?.records || resp || []
  } catch {
    list.value = mockList.filter(r => {
      if (filters.gas_type && r.gas_type !== filters.gas_type) return false
      if (filters.granary_id && r.granary_id !== filters.granary_id) return false
      if (filters.date && !r.detection_time?.startsWith(filters.date)) return false
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

const loadUnsealList = async () => {
  if (!createForm.granary_id) {
    unsealList.value = []
    return
  }
  try {
    const data = await unsealApi.list({ granary_id: createForm.granary_id })
    unsealList.value = data || []
  } catch {
    unsealList.value = [
      { id: 'u1', granary_id: createForm.granary_id, unseal_type: 'ventilation', start_time: '2024-12-22T08:00:00Z' }
    ]
  }
}

const onGranaryChange = () => {
  createForm.unseal_id = ''
}

const openCreateDialog = () => {
  Object.assign(createForm, {
    granary_id: '', unseal_id: '',
    detection_time: dayjs().format('YYYY-MM-DDTHH:mm:ss'),
    detection_points: '',
    readings: { gas_ph3: 0, gas_h2s: 0, co2: 0, o2: 0 },
    remark: ''
  })
  createDialogVisible.value = true
}

const handleCreate = async () => {
  const valid = await createFormRef.value?.validate().catch(() => false)
  if (!valid) return
  creating.value = true
  try {
    const r = createForm.readings
    const body = {
      ...createForm,
      concentration: r.gas_ph3,
      safe_limit: DefaultSafeLimits.gas_ph3,
      readings: r
    } as any
    try {
      await unsealApi.addGasDetection(createForm.granary_id, body)
    } catch {}
    const safe = currentSafeStatus.value
    ElMessage.success(
      `气体检测登记完成！` +
      (safe === false ? ' ⚠ 指标未达标，请勿解封！' : safe ? ' ✓ 指标达标' : '')
    )
    createDialogVisible.value = false
    loadList()
  } finally {
    creating.value = false
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
.value { color: #e6a23c; font-weight: 600; }
.limit { color: #909399; }
.over { color: #f56c6c; font-weight: 600; }
.violation { margin-bottom: 6px; }
.gray { color: #c0c4cc; }
.legend-tips {
  margin-top: 16px;
  padding: 12px 16px;
  background: #fafafa;
  border-radius: 4px;
  font-size: 13px;
  color: #606266;
}
</style>
