<template>
  <div class="page">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-select v-model="filters.type" placeholder="类型筛选" clearable style="width: 140px" @change="loadList">
              <el-option label="通风" value="ventilation" />
              <el-option label="解封" value="unseal" />
            </el-select>
            <el-select v-model="filters.granary_id" placeholder="仓房筛选" clearable filterable style="width: 180px" @change="loadList">
              <el-option v-for="g in granaries" :key="g.id" :label="`${g.code} - ${g.name}`" :value="g.id" />
            </el-select>
            <el-button type="primary" :icon="Refresh" @click="loadList">查询</el-button>
          </div>
          <el-button v-if="userStore.hasRole('admin','duty_officer','keeper')" type="success" :icon="Plus" @click="openCreateDialog">
            登记通风/解封
          </el-button>
        </div>
      </template>

      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column label="类型" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.unseal_type === 'ventilation' ? 'warning' : 'info'" effect="dark" size="small">
              {{ row.unseal_type === 'ventilation' ? '通风' : '解封' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="仓房" width="130">
          <template #default="{ row }">{{ row.granary?.name || '-' }}</template>
        </el-table-column>
        <el-table-column label="关联熏蒸方案" width="170">
          <template #default="{ row }">
            <el-tag v-if="row.fumigation_plan_id" type="danger" size="small" plain>{{ row.fumigation_plan_id }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="开始时间" width="170">
          <template #default="{ row }">{{ formatTime(row.start_time) }}</template>
        </el-table-column>
        <el-table-column label="结束时间" width="170">
          <template #default="{ row }">{{ formatTime(row.end_time) }}</template>
        </el-table-column>
        <el-table-column label="天气" width="100">
          <template #default="{ row }">{{ row.weather_condition || '-' }}</template>
        </el-table-column>
        <el-table-column label="安全状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_safe ? 'success' : 'warning'" size="small">
              {{ row.is_safe ? '安全' : '待检测' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="登记人" width="110">
          <template #default="{ row }">{{ row.recorder?.full_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetail(row)">详情</el-button>
            <el-button
              v-if="!row.end_time && userStore.hasRole('admin','duty_officer')"
              type="success"
              link
              size="small"
              @click="openCompleteDialog(row)"
            >完成登记</el-button>
            <el-button
              v-if="userStore.hasRole('admin','duty_officer')"
              type="warning"
              link
              size="small"
              @click="addGasDetection(row)"
            >气体检测</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="createDialogVisible" :title="createForm.unseal_type === 'ventilation' ? '登记通风' : '登记解封'" width="600px" destroy-on-close>
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="110px">
        <el-row :gutter="16">
          <el-col :span="24">
            <el-form-item label="操作类型">
              <el-radio-group v-model="createForm.unseal_type">
                <el-radio value="ventilation">通风</el-radio>
                <el-radio value="unseal">解封</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="仓房" prop="granary_id">
              <el-select v-model="createForm.granary_id" placeholder="选择仓房" filterable style="width: 100%">
                <el-option v-for="g in granaries" :key="g.id" :label="`${g.code} - ${g.name}`" :value="g.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="关联熏蒸方案">
              <el-select v-model="createForm.fumigation_plan_id" placeholder="选择（可选）" filterable clearable style="width: 100%">
                <el-option v-for="p in completedFumigations" :key="p.id" :label="p.plan_no" :value="p.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="开始时间">
              <el-date-picker v-model="createForm.start_time" type="datetime" style="width: 100%" value-format="YYYY-MM-DDTHH:mm:ss" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="天气情况">
              <el-select v-model="createForm.weather_condition" placeholder="请选择" clearable style="width: 100%">
                <el-option label="晴" value="晴" />
                <el-option label="多云" value="多云" />
                <el-option label="阴" value="阴" />
                <el-option label="雨" value="雨" />
                <el-option label="雪" value="雪" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
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

    <el-dialog v-model="completeDialogVisible" title="完成登记" width="560px" destroy-on-close>
      <div v-if="completeUnsealRow?.unseal_type === 'unseal'" style="margin-bottom: 16px">
        <el-alert type="error" :closable="false" show-icon>
          <template #title>
            解封前必须确认所有气体浓度已达到安全限值以下！请先进行气体检测。
          </template>
        </el-alert>
      </div>
      <el-form :model="completeForm" label-width="110px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="结束时间">
              <el-date-picker v-model="completeForm.end_time" type="datetime" style="width: 100%" value-format="YYYY-MM-DDTHH:mm:ss" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="是否安全达标">
              <el-switch v-model="completeForm.is_safe" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="最终气体数据">
              <el-input v-model="completeForm.final_gas_readings" type="textarea" :rows="2" placeholder="例：PH3: 0.2ppm，H2S: 未检出" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="completeForm.remark" type="textarea" :rows="3" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="completeDialogVisible = false">取消</el-button>
        <el-button type="success" :loading="completing" @click="handleComplete">确认完成</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="gasDialogVisible" title="登记气体检测" width="560px" destroy-on-close>
      <el-form :model="gasForm" label-width="110px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="气体类型">
              <el-select v-model="gasForm.gas_type" style="width: 100%">
                <el-option label="磷化氢 PH₃" value="gas_ph3" />
                <el-option label="硫化氢 H₂S" value="gas_h2s" />
                <el-option label="二氧化碳 CO₂" value="co2" />
                <el-option label="氧气 O₂" value="o2" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="浓度(ppm)">
              <el-input-number v-model="gasForm.concentration" :precision="3" :step="0.1" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="安全限值(ppm)">
              <el-input-number v-model="gasForm.safe_limit" :precision="3" :step="0.1" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="检测时间">
              <el-date-picker v-model="gasForm.detection_time" type="datetime" style="width: 100%" value-format="YYYY-MM-DDTHH:mm:ss" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="检测点">
              <el-input v-model="gasForm.detection_points" placeholder="例：仓内中心、北侧、南侧、东侧、西侧" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="gasForm.remark" type="textarea" :rows="2" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="gasDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="gasSubmitting" @click="handleGasDetection">登记</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Refresh, Plus } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import type { Granary, UnsealRecord, UnsealType, FumigationPlan } from '@/types'
import { granaryApi, unsealApi, fumigationApi } from '@/api'
import dayjs from 'dayjs'

const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const list = ref<UnsealRecord[]>([])
const granaries = ref<Granary[]>([])
const completedFumigations = ref<FumigationPlan[]>([])

const filters = reactive({ type: '' as UnsealType | '', granary_id: '' })

const createDialogVisible = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = reactive({
  granary_id: '', fumigation_plan_id: '',
  unseal_type: 'ventilation' as UnsealType,
  start_time: dayjs().format('YYYY-MM-DDTHH:mm:ss'),
  weather_condition: '', remark: ''
})
const createRules: FormRules = {
  granary_id: [{ required: true, message: '请选择仓房', trigger: 'change' }]
}

const completeDialogVisible = ref(false)
const completing = ref(false)
const completeUnsealRow = ref<UnsealRecord>()
const completeForm = reactive({
  end_time: dayjs().format('YYYY-MM-DDTHH:mm:ss'),
  is_safe: false, final_gas_readings: '', remark: ''
})

const gasDialogVisible = ref(false)
const gasSubmitting = ref(false)
const gasGranaryId = ref('')
const gasUnsealId = ref('')
const defaultLimits: Record<string, number> = { gas_ph3: 0.3, gas_h2s: 10, co2: 5000, o2: 195000 }
const gasForm = reactive({
  unseal_id: '',
  gas_type: 'gas_ph3',
  concentration: 0,
  safe_limit: 0.3,
  detection_time: dayjs().format('YYYY-MM-DDTHH:mm:ss'),
  detection_points: '',
  remark: ''
})

const safeStatusLabel = computed(() => {
  if (gasForm.safe_limit > 0) {
    const safe = gasForm.concentration <= gasForm.safe_limit
    return safe ? '已达标' : '未达标'
  }
  return ''
})

const formatTime = (t?: string) => t ? dayjs(t).format('YYYY-MM-DD HH:mm') : '-'

const loadList = async () => {
  loading.value = true
  try {
    const data = await unsealApi.list({ type: filters.type || undefined, granary_id: filters.granary_id || undefined })
    list.value = data
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

const loadGranaries = async () => {
  try {
    granaries.value = await granaryApi.list()
  } catch {
    granaries.value = []
  }
}

const loadCompletedFumigations = async () => {
  try {
    completedFumigations.value = await fumigationApi.listPlans({ status: 'completed' })
  } catch {
    completedFumigations.value = []
  }
}

const viewDetail = (row: UnsealRecord) => {
  ElMessageBox.alert(
    `类型: ${row.unseal_type === 'ventilation' ? '通风' : '解封'}\n` +
    `仓房: ${row.granary?.name || '-'}\n` +
    `开始: ${formatTime(row.start_time)}\n` +
    `结束: ${formatTime(row.end_time)}\n` +
    `天气: ${row.weather_condition || '-'}\n` +
    `状态: ${row.is_safe ? '安全' : '待检测'}\n` +
    `备注: ${row.remark || '-'}`,
    '详情', { confirmButtonText: '关闭' }
  )
}

const openCreateDialog = () => {
  Object.assign(createForm, {
    granary_id: '', fumigation_plan_id: '', unseal_type: 'ventilation',
    start_time: dayjs().format('YYYY-MM-DDTHH:mm:ss'),
    weather_condition: '', remark: ''
  })
  createDialogVisible.value = true
}

const handleCreate = async () => {
  const valid = await createFormRef.value?.validate().catch(() => false)
  if (!valid) return
  creating.value = true
  try {
    await unsealApi.create(createForm)
    ElMessage.success('登记成功')
    createDialogVisible.value = false
    loadList()
  } catch {
  } finally {
    creating.value = false
  }
}

const openCompleteDialog = (row: UnsealRecord) => {
  completeUnsealRow.value = row
  Object.assign(completeForm, {
    end_time: dayjs().format('YYYY-MM-DDTHH:mm:ss'),
    is_safe: false, final_gas_readings: '', remark: ''
  })
  completeDialogVisible.value = true
}

const handleComplete = async () => {
  if (!completeUnsealRow.value) return
  completing.value = true
  try {
    await unsealApi.complete(completeUnsealRow.value.id, completeForm)
    ElMessage.success('登记完成')
    completeDialogVisible.value = false
    loadList()
  } catch {
  } finally {
    completing.value = false
  }
}

const addGasDetection = (row: UnsealRecord) => {
  gasGranaryId.value = row.granary_id
  gasUnsealId.value = row.id
  Object.assign(gasForm, {
    unseal_id: row.id, gas_type: 'gas_ph3', concentration: 0,
    safe_limit: 0.3,
    detection_time: dayjs().format('YYYY-MM-DDTHH:mm:ss'),
    detection_points: '', remark: ''
  })
  gasDialogVisible.value = true
}

const handleGasDetection = async () => {
  gasSubmitting.value = true
  try {
    const body: any = { ...gasForm }
    if (!body.unseal_id) delete body.unseal_id
    await unsealApi.addGasDetection(gasGranaryId.value, body)
    const safe = gasForm.concentration <= gasForm.safe_limit
    ElMessage.success(`气体检测登记完成，浓度${gasForm.concentration} ppm，${safe ? '已达标' : '未达标'}，安全限值 ${gasForm.safe_limit} ppm`)
    gasDialogVisible.value = false
  } catch {
  } finally {
    gasSubmitting.value = false
  }
}

onMounted(() => {
  loadList()
  loadGranaries()
  loadCompletedFumigations()
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
</style>
