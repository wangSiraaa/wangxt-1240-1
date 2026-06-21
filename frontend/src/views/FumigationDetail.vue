<template>
  <div class="page">
    <el-page-header @back="$router.back()" :content="detail?.plan_title || '熏蒸方案详情'" class="page-header">
      <template #extra>
        <el-tag v-if="detail" :type="FumigationStatusColors[detail.status] as any" effect="dark" size="large">
          {{ FumigationStatusLabels[detail.status] }}
        </el-tag>
      </template>
    </el-page-header>

    <el-row :gutter="20" style="margin-top: 16px">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header><div class="card-title">方案信息</div></template>
          <el-descriptions v-if="detail" :column="2" border size="small">
            <el-descriptions-item label="方案编号">{{ detail.plan_no }}</el-descriptions-item>
            <el-descriptions-item label="仓房">{{ detail.granary?.name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="药剂类型">{{ detail.chemical_type || '-' }}</el-descriptions-item>
            <el-descriptions-item label="药剂名称">{{ detail.chemical_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="用药量">{{ detail.dosage || '-' }} {{ detail.dosage_unit || '' }}</el-descriptions-item>
            <el-descriptions-item label="目标浓度">{{ detail.target_concentration || '-' }} ppm</el-descriptions-item>
            <el-descriptions-item label="计划开始">{{ formatTime(detail.plan_start_time) }}</el-descriptions-item>
            <el-descriptions-item label="计划结束">{{ formatTime(detail.plan_end_time) }}</el-descriptions-item>
            <el-descriptions-item label="预期密封">{{ detail.expected_seal_hours || '-' }} 小时</el-descriptions-item>
            <el-descriptions-item label="检测间隔">{{ detail.detection_interval_hours || 4 }} 小时</el-descriptions-item>
            <el-descriptions-item label="创建人">{{ detail.creator?.full_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="审批人">{{ detail.approver?.full_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="审批时间">{{ formatTime(detail.approved_at) }}</el-descriptions-item>
            <el-descriptions-item label="下次检测">{{ formatTime(detail.next_detection_time) }}</el-descriptions-item>
            <el-descriptions-item label="安全员确认">
              <el-tag v-if="detail.safety_confirmed" type="success">已确认</el-tag>
              <el-tag v-else type="warning">待确认</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="确认人">{{ detail.safety_confirmed_by ? (getConfirmerName()) : '-' }}</el-descriptions-item>
            <el-descriptions-item label="确认时间">{{ formatTime(detail.safety_confirmed_at) }}</el-descriptions-item>
            <el-descriptions-item label="熏蒸原因" :span="2">{{ detail.reason || '-' }}</el-descriptions-item>
            <el-descriptions-item label="审批意见" :span="2">{{ detail.approval_remark || '-' }}</el-descriptions-item>
            <el-descriptions-item label="确认备注" :span="2">{{ detail.safety_confirm_remark || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>执行状态</span>
              <el-tag v-if="detail?.people_cleared" type="success">人员已清场</el-tag>
              <el-tag v-else type="danger">人员未清场</el-tag>
            </div>
          </template>
          <el-steps :active="stepIndex" direction="vertical" finish-status="success" process-status="warning">
            <el-step title="创建方案" :description="detail?.created_at ? formatTime(detail.created_at) : '未完成'" />
            <el-step title="提交审批" :description="canSubmit ? '可提交审批' : '已提交'" />
            <el-step title="审批通过" :description="detail?.approved_at ? formatTime(detail.approved_at) : (detail?.status === 'rejected' ? '已驳回' : '待审批')" />
            <el-step title="确认清场" :description="detail?.people_cleared ? (detail.people_cleared_time ? formatTime(detail.people_cleared_time) : '已确认') : '未清场'" />
            <el-step title="投药执行" :description="detail?.status === 'in_progress' || detail?.status === 'completed' ? '已开始' : '未开始'" />
            <el-step title="完成熏蒸" :description="detail?.status === 'completed' ? '已完成' : '未完成'" />
            <el-step title="通风复测" :description="ventilationStatus" />
            <el-step title="安全员确认" :description="safetyConfirmStatus" />
            <el-step title="解封" :description="canUnseal ? '可进行解封' : '等待安全员确认'" />
          </el-steps>

          <div style="margin-top: 20px" v-if="detail?.status === 'completed' && detail?.next_detection_time && isDetectionOverdue">
            <el-alert
              :title="`已超过下次检测时间（${formatTime(detail.next_detection_time)}），请立即进行气体检测！`"
              type="error"
              :closable="false"
              show-icon
            />
          </div>

          <div style="margin-top: 20px" v-else-if="detail?.status === 'completed' && detail?.next_detection_time && !detail.safety_confirmed">
            <el-alert
              :title="`下次气体检测时间：${formatTime(detail.next_detection_time)}，请按时检测并由安全员确认达标`"
              type="warning"
              :closable="false"
              show-icon
            >
              <template #default>
                <el-button type="primary" link @click="goUnseal">去通风检测</el-button>
              </template>
            </el-alert>
          </div>

          <div style="margin-top: 20px" v-else-if="detail?.status === 'completed' && !detail?.next_detection_time">
            <el-alert
              title="熏蒸已完成，请先登记通风，系统将自动提示下次气体检测时间"
              type="success"
              :closable="false"
              show-icon
            >
              <template #default>
                <el-button type="primary" link @click="goUnseal">立即登记通风</el-button>
              </template>
            </el-alert>
          </div>

          <div style="margin-top: 20px" v-if="detail?.status === 'completed' && !detail?.safety_confirmed && userStore.hasRole('safety_officer', 'admin')">
            <el-button type="success" :icon="Check" @click="openSafetyConfirmDialog" style="width: 100%">
              安全员确认达标
            </el-button>
          </div>

          <div style="margin-top: 20px" v-if="detail?.status === 'completed' && detail?.safety_confirmed">
            <el-alert
              title="安全员已确认气体检测达标，可以进行解封"
              type="success"
              :closable="false"
              show-icon
            >
              <template #default>
                <el-button type="primary" link @click="goUnseal">去解封</el-button>
              </template>
            </el-alert>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row style="margin-top: 20px">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header><div class="card-title">执行记录</div></template>
          <el-table :data="executions" size="small" empty-text="暂无执行记录">
            <el-table-column label="开始时间" width="180">
              <template #default="{ row }">{{ formatTime(row.actual_start_time) }}</template>
            </el-table-column>
            <el-table-column label="结束时间" width="180">
              <template #default="{ row }">{{ formatTime(row.actual_end_time) }}</template>
            </el-table-column>
            <el-table-column prop="chemical_actual_dosage" label="实际用药量" width="140" />
            <el-table-column prop="weather_during" label="天气" width="120" />
            <el-table-column prop="remark" label="备注" show-overflow-tooltip />
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="safetyConfirmDialogVisible" title="安全员确认达标" width="500px" destroy-on-close>
      <el-form :model="safetyConfirmForm" label-width="100px">
        <el-form-item label="是否达标">
          <el-radio-group v-model="safetyConfirmForm.confirmed">
            <el-radio :value="true">达标，确认</el-radio>
            <el-radio :value="false">不达标，需继续通风</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="safetyConfirmForm.remark" type="textarea" :rows="3" placeholder="请输入确认备注（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="safetyConfirmDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="safetyConfirming" @click="handleSafetyConfirm">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Check } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import type { FumigationPlan, FumigationExecution } from '@/types'
import { FumigationStatusLabels, FumigationStatusColors } from '@/types'
import { fumigationApi } from '@/api'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const planId = computed(() => route.params.id as string)

const detail = ref<FumigationPlan>()
const executions = ref<FumigationExecution[]>([])

const safetyConfirmDialogVisible = ref(false)
const safetyConfirming = ref(false)
const safetyConfirmForm = reactive({
  confirmed: true,
  remark: ''
})

const stepIndex = computed(() => {
  if (!detail.value) return 0
  const s = detail.value.status
  if (s === 'draft') return 1
  if (s === 'pending_approval') return 2
  if (s === 'rejected') return 2
  if (s === 'approved' && !detail.value.people_cleared) return 3
  if (s === 'approved' && detail.value.people_cleared) return 4
  if (s === 'in_progress') return 5
  if (s === 'completed') {
    if (detail.value.next_detection_time) {
      if (detail.value.safety_confirmed) {
        return 9
      }
      return 7
    }
    return 6
  }
  return 0
})

const ventilationStatus = computed(() => {
  if (!detail.value) return ''
  if (detail.value.status !== 'completed') return '等待完成熏蒸'
  if (detail.value.next_detection_time) return `已通风，下次检测：${formatTime(detail.value.next_detection_time)}`
  return '待登记通风'
})

const safetyConfirmStatus = computed(() => {
  if (!detail.value) return ''
  if (detail.value.status !== 'completed') return '等待完成熏蒸'
  if (!detail.value.next_detection_time) return '待通风登记'
  if (detail.value.safety_confirmed) return `已确认 (${formatTime(detail.value.safety_confirmed_at)})`
  return '待安全员确认'
})

const isDetectionOverdue = computed(() => {
  if (!detail.value?.next_detection_time) return false
  return dayjs().isAfter(dayjs(detail.value.next_detection_time))
})

const canSubmit = computed(() => detail.value?.status === 'draft' || detail.value?.status === 'rejected')
const canUnseal = computed(() => detail.value?.status === 'completed' && detail.value?.safety_confirmed)

const formatTime = (t?: string) => t ? dayjs(t).format('YYYY-MM-DD HH:mm') : '-'

const getConfirmerName = () => {
  return detail.value?.safety_confirmed_by || '-'
}

const loadDetail = async () => {
  try {
    detail.value = await fumigationApi.getPlan(planId.value)
  } catch {
  }
}

const loadExecutions = async () => {
  try {
    executions.value = await fumigationApi.listExecutions({ plan_id: planId.value })
  } catch {
    executions.value = []
  }
}

const goUnseal = () => {
  router.push('/unseal')
}

const openSafetyConfirmDialog = () => {
  safetyConfirmForm.confirmed = true
  safetyConfirmForm.remark = ''
  safetyConfirmDialogVisible.value = true
}

const handleSafetyConfirm = async () => {
  safetyConfirming.value = true
  try {
    await fumigationApi.safetyConfirm(planId.value, safetyConfirmForm)
    ElMessage.success(safetyConfirmForm.confirmed ? '已确认达标' : '已标记为不达标')
    safetyConfirmDialogVisible.value = false
    loadDetail()
  } catch {
  } finally {
    safetyConfirming.value = false
  }
}

onMounted(() => {
  loadDetail()
  loadExecutions()
})
</script>

<style scoped>
.page { width: 100%; }
.page-header {
  background: #fff;
  padding: 16px 20px;
  border-radius: 4px;
  margin: -4px;
}
.card-title { font-weight: 600; }
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
