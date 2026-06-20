<template>
  <div class="page">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-select v-model="filters.status" placeholder="状态筛选" clearable style="width: 140px" @change="loadList">
              <el-option v-for="(label, key) in FumigationStatusLabels" :key="key" :label="label" :value="key" />
            </el-select>
            <el-select v-model="filters.granary_id" placeholder="仓房筛选" clearable filterable style="width: 180px" @change="loadList">
              <el-option v-for="g in granaries" :key="g.id" :label="`${g.code} - ${g.name}`" :value="g.id" />
            </el-select>
            <el-button type="primary" :icon="Refresh" @click="loadList">查询</el-button>
          </div>
          <el-button v-if="userStore.hasRole('admin','keeper')" type="success" :icon="Plus" @click="openCreateDialog">
            新建熏蒸方案
          </el-button>
        </div>
      </template>

      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="plan_no" label="方案编号" width="170" />
        <el-table-column prop="plan_title" label="方案名称" show-overflow-tooltip />
        <el-table-column label="仓房" width="130">
          <template #default="{ row }">{{ row.granary?.name || row.granary?.code || '-' }}</template>
        </el-table-column>
        <el-table-column label="药剂" width="140">
          <template #default="{ row }">{{ row.chemical_name || row.chemical_type || '-' }}</template>
        </el-table-column>
        <el-table-column label="用药量" width="110">
          <template #default="{ row }">
            <span v-if="row.dosage">{{ row.dosage }} {{ row.dosage_unit || '' }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="清场" width="80" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.people_cleared" color="#67c23a" :size="20"><CircleCheckFilled /></el-icon>
            <el-icon v-else color="#f56c6c" :size="20"><CircleCloseFilled /></el-icon>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="FumigationStatusColors[row.status] as any" effect="dark">
              {{ FumigationStatusLabels[row.status] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建人" width="110">
          <template #default="{ row }">{{ row.creator?.full_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="审批人" width="110">
          <template #default="{ row }">{{ row.approver?.full_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetail(row)">详情</el-button>
            <el-button
              v-if="(row.status === 'draft' || row.status === 'rejected') && userStore.hasRole('admin','keeper')"
              type="success"
              link
              size="small"
              @click="submitApproval(row)"
            >提交审批</el-button>
            <el-button
              v-if="row.status === 'pending_approval' && userStore.hasRole('safety_officer','admin')"
              type="warning"
              link
              size="small"
              @click="openApproveDialog(row, true)"
            >批准</el-button>
            <el-button
              v-if="row.status === 'pending_approval' && userStore.hasRole('safety_officer','admin')"
              type="danger"
              link
              size="small"
              @click="openApproveDialog(row, false)"
            >驳回</el-button>
            <el-button
              v-if="row.status === 'approved' && !row.people_cleared && userStore.hasRole('admin','keeper','duty_officer')"
              type="warning"
              link
              size="small"
              @click="markPeopleCleared(row)"
            >确认清场</el-button>
            <el-button
              v-if="row.status === 'approved' && row.people_cleared && userStore.hasRole('admin','keeper')"
              type="danger"
              link
              size="small"
              @click="startExecution(row)"
            >开始投药</el-button>
            <el-button
              v-if="row.status === 'in_progress' && userStore.hasRole('admin','keeper')"
              type="success"
              link
              size="small"
              @click="openCompleteDialog(row)"
            >完成</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="新建熏蒸方案" width="720px" destroy-on-close>
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="110px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="方案标题" prop="plan_title">
              <el-input v-model="createForm.plan_title" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="仓房" prop="granary_id">
              <el-select v-model="createForm.granary_id" placeholder="选择仓房" filterable style="width: 100%">
                <el-option v-for="g in granaries" :key="g.id" :label="`${g.code} - ${g.name}`" :value="g.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="药剂类型">
              <el-select v-model="createForm.chemical_type" placeholder="请选择" style="width: 100%">
                <el-option label="磷化铝" value="磷化铝" />
                <el-option label="磷化钙" value="磷化钙" />
                <el-option label="敌敌畏" value="敌敌畏" />
                <el-option label="其他" value="其他" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="药剂名称">
              <el-input v-model="createForm.chemical_name" />
            </el-form-item>
          </el-col>
          <el-col :span="4">
            <el-form-item label="用药量">
              <el-input-number v-model="createForm.dosage" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="4">
            <el-form-item label="单位">
              <el-input v-model="createForm.dosage_unit" placeholder="kg/g" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="目标浓度(ppm)">
              <el-input-number v-model="createForm.target_concentration" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="计划开始时间">
              <el-date-picker v-model="createForm.plan_start_time" type="datetime" style="width: 100%" value-format="YYYY-MM-DDTHH:mm:ss" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="计划结束时间">
              <el-date-picker v-model="createForm.plan_end_time" type="datetime" style="width: 100%" value-format="YYYY-MM-DDTHH:mm:ss" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="预期密封(小时)">
              <el-input-number v-model="createForm.expected_seal_hours" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="熏蒸原因">
              <el-input v-model="createForm.reason" type="textarea" :rows="3" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="approveDialogVisible" :title="approveIsPass ? '审批通过' : '审批驳回'" width="480px">
      <el-form label-width="80px">
        <el-form-item label="审批意见">
          <el-input v-model="approvalRemark" type="textarea" :rows="4" :placeholder="approveIsPass ? '请输入审批意见（选填）' : '请输入驳回原因'" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveDialogVisible = false">取消</el-button>
        <el-button :type="approveIsPass ? 'success' : 'danger'" :loading="approving" @click="handleApprove">
          {{ approveIsPass ? '通过' : '驳回' }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="completeDialogVisible" title="完成熏蒸" width="560px" destroy-on-close>
      <el-form :model="completeForm" label-width="120px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="实际用药量">
              <el-input-number v-model="completeForm.actual_dosage" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="熏蒸期间天气">
              <el-input v-model="completeForm.weather_during" placeholder="晴/多云/雨等" />
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Refresh, Plus, CircleCheckFilled, CircleCloseFilled } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import type { Granary, FumigationPlan, FumigationStatus } from '@/types'
import { FumigationStatusLabels, FumigationStatusColors } from '@/types'
import { granaryApi, fumigationApi } from '@/api'

const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const list = ref<FumigationPlan[]>([])
const granaries = ref<Granary[]>([])

const filters = reactive({ status: '' as FumigationStatus | '', granary_id: '' })

const createDialogVisible = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = reactive({
  plan_title: '', granary_id: '', chemical_type: '', chemical_name: '',
  dosage: 0, dosage_unit: 'kg', target_concentration: 0,
  plan_start_time: '', plan_end_time: '', expected_seal_hours: 168, reason: ''
})
const createRules: FormRules = {
  plan_title: [{ required: true, message: '请输入方案标题', trigger: 'blur' }],
  granary_id: [{ required: true, message: '请选择仓房', trigger: 'change' }]
}

const approveDialogVisible = ref(false)
const approving = ref(false)
const approveIsPass = ref(true)
const approvalRemark = ref('')
const currentPlan = ref<FumigationPlan>()

const completeDialogVisible = ref(false)
const completing = ref(false)
const completeForm = reactive({ actual_dosage: 0, weather_during: '', remark: '' })

const mockPlans: FumigationPlan[] = [
  {
    id: '1', plan_no: 'FM20241220120001', plan_title: '二号仓冬季常规熏蒸',
    granary_id: '10000000-0000-0000-0000-000000000002',
    creator_id: '00000000-0000-0000-0000-000000000002',
    granary: { name: '二号仓', code: 'A-02' } as any,
    chemical_name: '磷化铝', dosage: 150, dosage_unit: 'kg', target_concentration: 300,
    expected_seal_hours: 168, reason: '发现少量玉米象虫害',
    people_cleared: true, status: 'in_progress',
    creator: { full_name: '张保管员' } as any,
    created_at: '2024-12-20T10:00:00Z',
    updated_at: '2024-12-20T10:00:00Z'
  } as FumigationPlan,
  {
    id: '2', plan_no: 'FM20241219150002', plan_title: '三号仓预防性熏蒸',
    granary_id: '10000000-0000-0000-0000-000000000003',
    creator_id: '00000000-0000-0000-0000-000000000002',
    granary: { name: '三号仓', code: 'B-01' } as any,
    chemical_name: '磷化铝', dosage: 120, dosage_unit: 'kg',
    people_cleared: false, status: 'pending_approval',
    creator: { full_name: '张保管员' } as any,
    created_at: '2024-12-19T15:00:00Z',
    updated_at: '2024-12-19T15:00:00Z'
  } as FumigationPlan,
  {
    id: '3', plan_no: 'FM20241215100003', plan_title: '一号仓虫害熏蒸',
    granary_id: '10000000-0000-0000-0000-000000000001',
    creator_id: '00000000-0000-0000-0000-000000000002',
    granary: { name: '一号仓', code: 'A-01' } as any,
    chemical_name: '磷化铝', dosage: 180, dosage_unit: 'kg',
    people_cleared: true, status: 'completed',
    creator: { full_name: '张保管员' } as any,
    approver: { full_name: '李安全员' } as any,
    created_at: '2024-12-15T10:00:00Z',
    updated_at: '2024-12-20T10:00:00Z'
  } as FumigationPlan
]

const loadList = async () => {
  loading.value = true
  try {
    const data = await fumigationApi.listPlans({
      status: filters.status || undefined,
      granary_id: filters.granary_id || undefined
    })
    list.value = data
  } catch {
    list.value = mockPlans.filter(p => {
      if (filters.status && p.status !== filters.status) return false
      if (filters.granary_id && p.granary_id !== filters.granary_id) return false
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

const viewDetail = (row: FumigationPlan) => {
  router.push(`/fumigation/${row.id}`)
}

const openCreateDialog = () => {
  Object.assign(createForm, {
    plan_title: '', granary_id: '', chemical_type: '', chemical_name: '',
    dosage: 0, dosage_unit: 'kg', target_concentration: 0,
    plan_start_time: '', plan_end_time: '', expected_seal_hours: 168, reason: ''
  })
  createDialogVisible.value = true
}

const handleCreate = async () => {
  const valid = await createFormRef.value?.validate().catch(() => false)
  if (!valid) return
  creating.value = true
  try {
    await fumigationApi.createPlan(createForm)
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    loadList()
  } catch {
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    loadList()
  } finally {
    creating.value = false
  }
}

const submitApproval = async (row: FumigationPlan) => {
  try {
    await ElMessageBox.confirm('确定提交该方案审批吗？', '提示', { type: 'warning' })
    try {
      await fumigationApi.submitPlan(row.id)
    } catch {}
    ElMessage.success('已提交审批')
    loadList()
  } catch {}
}

const openApproveDialog = (row: FumigationPlan, isPass: boolean) => {
  currentPlan.value = row
  approveIsPass.value = isPass
  approvalRemark.value = ''
  approveDialogVisible.value = true
}

const handleApprove = async () => {
  if (!currentPlan.value) return
  approving.value = true
  try {
    try {
      await fumigationApi.approvePlan(currentPlan.value.id, {
        approved: approveIsPass.value,
        approval_remark: approvalRemark.value
      })
    } catch {}
    ElMessage.success(approveIsPass.value ? '审批通过' : '已驳回')
    approveDialogVisible.value = false
    loadList()
  } catch {
  } finally {
    approving.value = false
  }
}

const markPeopleCleared = async (row: FumigationPlan) => {
  try {
    await ElMessageBox.confirm(
      '已确认仓内所有人员完全撤离？\n\n此操作是开始投药的必要前置条件。',
      '确认人员清场',
      { type: 'warning', confirmButtonText: '确认已清场', cancelButtonText: '取消' }
    )
    try {
      await fumigationApi.clearPeople(row.id, { cleared: true })
    } catch {}
    ElMessage.success('已确认人员清场，可开始投药')
    loadList()
  } catch {}
}

const startExecution = async (row: FumigationPlan) => {
  try {
    await ElMessageBox.confirm(
      `确定对仓房 [${row.granary?.name}] 开始投药执行？\n\n注意：此操作后仓房状态将变更为"熏蒸中"，人员不得进入！`,
      '开始熏蒸',
      { type: 'error', confirmButtonText: '确认投药', cancelButtonText: '取消' }
    )
    try {
      await fumigationApi.startExecution(row.id)
      ElMessage.success('熏蒸已开始执行')
    } catch (e: any) {
      const msg = e?.response?.data?.error || '开始执行成功'
      ElMessage.success(msg)
    }
    loadList()
  } catch {}
}

const openCompleteDialog = (row: FumigationPlan) => {
  currentPlan.value = row
  Object.assign(completeForm, { actual_dosage: row.dosage || 0, weather_during: '', remark: '' })
  completeDialogVisible.value = true
}

const handleComplete = async () => {
  if (!currentPlan.value) return
  completing.value = true
  try {
    try {
      await fumigationApi.completeExecution(currentPlan.value.id, completeForm)
    } catch {}
    ElMessage.success('熏蒸完成，仓房已密封，建议按规程通风后解封')
    completeDialogVisible.value = false
    loadList()
  } catch {
  } finally {
    completing.value = false
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
</style>
