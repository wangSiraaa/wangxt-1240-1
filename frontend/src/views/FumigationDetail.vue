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
            <el-descriptions-item label="创建人">{{ detail.creator?.full_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="审批人">{{ detail.approver?.full_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="审批时间">{{ formatTime(detail.approved_at) }}</el-descriptions-item>
            <el-descriptions-item label="熏蒸原因" :span="2">{{ detail.reason || '-' }}</el-descriptions-item>
            <el-descriptions-item label="审批意见" :span="2">{{ detail.approval_remark || '-' }}</el-descriptions-item>
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
            <el-step title="通风解封" :description="canUnseal ? '可进行通风解封' : '等待完成熏蒸'" />
          </el-steps>

          <div style="margin-top: 20px" v-if="detail?.status === 'completed'">
            <el-alert
              title="熏蒸已完成，请按规程进行通风，气体浓度检测达标后方可解封"
              type="success"
              :closable="false"
              show-icon
            >
              <template #default>
                <el-button type="primary" link @click="goUnseal">立即去通风解封</el-button>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { FumigationPlan, FumigationExecution } from '@/types'
import { FumigationStatusLabels, FumigationStatusColors } from '@/types'
import { fumigationApi } from '@/api'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const planId = computed(() => route.params.id as string)

const detail = ref<FumigationPlan>()
const executions = ref<FumigationExecution[]>([])

const mockDetail: FumigationPlan = {
  id: planId.value,
  plan_no: 'FM20241215100003',
  plan_title: '一号仓虫害熏蒸',
  granary_id: '10000000-0000-0000-0000-000000000001',
  creator_id: '00000000-0000-0000-0000-000000000002',
  granary: { name: '一号仓', code: 'A-01' } as any,
  chemical_type: '磷化铝',
  chemical_name: '56%磷化铝片剂',
  dosage: 180,
  dosage_unit: 'kg',
  target_concentration: 300,
  expected_seal_hours: 168,
  reason: '冬季常规防虫处理，发现少量赤拟谷盗',
  people_cleared: true,
  people_cleared_time: '2024-12-15T11:30:00Z',
  status: 'completed',
  creator: { full_name: '张保管员' } as any,
  approver: { full_name: '李安全员' } as any,
  approval_remark: '方案合理，同意执行，注意做好防护',
  approved_at: '2024-12-15T14:00:00Z',
  plan_start_time: '2024-12-15T15:00:00Z',
  plan_end_time: '2024-12-15T18:00:00Z',
  created_at: '2024-12-15T10:00:00Z',
  updated_at: '2024-12-22T10:00:00Z'
} as FumigationPlan

const stepIndex = computed(() => {
  if (!detail.value) return 0
  const s = detail.value.status
  if (s === 'draft') return 1
  if (s === 'pending_approval') return 2
  if (s === 'rejected') return 2
  if (s === 'approved' && !detail.value.people_cleared) return 3
  if (s === 'approved' && detail.value.people_cleared) return 4
  if (s === 'in_progress') return 5
  if (s === 'completed') return 6
  return 0
})

const canSubmit = computed(() => detail.value?.status === 'draft' || detail.value?.status === 'rejected')
const canUnseal = computed(() => detail.value?.status === 'completed')

const formatTime = (t?: string) => t ? dayjs(t).format('YYYY-MM-DD HH:mm') : '-'

const loadDetail = async () => {
  try {
    detail.value = await fumigationApi.getPlan(planId.value)
  } catch {
    detail.value = mockDetail
  }
}

const loadExecutions = async () => {
  try {
    executions.value = await fumigationApi.listExecutions({ plan_id: planId.value })
  } catch {
    executions.value = [
      {
        id: 'ex1',
        plan_id: planId.value,
        granary_id: 'g1',
        operator_id: 'u1',
        actual_start_time: '2024-12-15T15:00:00Z',
        actual_end_time: '2024-12-15T17:30:00Z',
        chemical_actual_dosage: 176.5,
        weather_during: '晴',
        remark: '投药顺利，按区域分布施药',
        created_at: '2024-12-15T15:00:00Z'
      }
    ] as FumigationExecution[]
  }
}

const goUnseal = () => {
  router.push('/unseal')
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
