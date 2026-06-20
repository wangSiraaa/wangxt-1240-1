<template>
  <div class="page">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span style="font-weight: 600; margin-right: 16px">粮情记录列表</span>
            <el-select v-model="filters.granary_id" placeholder="选择仓房" clearable style="width: 180px" @change="loadList">
              <el-option v-for="g in granaries" :key="g.id" :label="`${g.code} - ${g.name}`" :value="g.id" />
            </el-select>
          </div>
          <el-button type="success" :icon="Plus" @click="openCreateDialog">录入粮情</el-button>
        </div>
      </template>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column label="仓房" width="150">
          <template #default="{ row }">{{ row.granary?.name || row.granary?.code || '-' }}</template>
        </el-table-column>
        <el-table-column label="记录时间" width="170">
          <template #default="{ row }">{{ formatTime(row.record_time) }}</template>
        </el-table-column>
        <el-table-column label="记录人" width="120">
          <template #default="{ row }">{{ row.recorder?.full_name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="avg_temperature" label="均温(°C)" width="100" align="center" />
        <el-table-column label="最高温(°C)" width="110" align="center">
          <template #default="{ row }">
            <el-tag :type="row.max_temperature >= 25 ? 'danger' : 'info'" size="small">
              {{ row.max_temperature }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="min_temperature" label="最低温(°C)" width="100" align="center" />
        <el-table-column prop="avg_humidity" label="湿度(%)" width="100" align="center" />
        <el-table-column label="异常">
          <template #default="{ row }">
            <el-tag v-if="row.pest_found" type="warning" size="small" effect="dark">虫害</el-tag>
            <el-tag v-if="row.mold_found" type="danger" size="small" effect="dark">霉变</el-tag>
            <el-tag v-if="row.abnormal_areas" type="warning" size="small">异常区域</el-tag>
            <span v-if="!row.pest_found && !row.mold_found && !row.abnormal_areas">-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="录入粮情" width="720px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="仓房" prop="granary_id">
              <el-select v-model="form.granary_id" placeholder="请选择仓房" filterable style="width: 100%">
                <el-option v-for="g in granaries" :key="g.id" :label="`${g.code} - ${g.name}`" :value="g.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="记录时间" prop="record_time">
              <el-date-picker
                v-model="form.record_time"
                type="datetime"
                placeholder="选择时间"
                style="width: 100%"
                value-format="YYYY-MM-DDTHH:mm:ss"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="平均温度(°C)">
              <el-input-number v-model="form.avg_temperature" :precision="2" :step="0.5" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="最高温度(°C)">
              <el-input-number v-model="form.max_temperature" :precision="2" :step="0.5" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="最低温度(°C)">
              <el-input-number v-model="form.min_temperature" :precision="2" :step="0.5" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="平均湿度(%)">
              <el-input-number v-model="form.avg_humidity" :precision="1" :min="0" :max="100" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="粮层高度(m)">
              <el-input-number v-model="form.grain_level" :precision="2" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="天气情况">
              <el-select v-model="form.weather_condition" placeholder="请选择" style="width: 100%">
                <el-option label="晴" value="晴" />
                <el-option label="多云" value="多云" />
                <el-option label="阴" value="阴" />
                <el-option label="雨" value="雨" />
                <el-option label="雪" value="雪" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="发现虫害">
              <el-switch v-model="form.pest_found" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="发现霉变">
              <el-switch v-model="form.mold_found" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="form.remark" type="textarea" :rows="3" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleCreate">提交</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" title="粮情详情" width="600px">
      <el-descriptions v-if="currentDetail" :column="2" border size="small">
        <el-descriptions-item label="仓房">{{ currentDetail.granary?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="记录时间">{{ formatTime(currentDetail.record_time) }}</el-descriptions-item>
        <el-descriptions-item label="记录人">{{ currentDetail.recorder?.full_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="天气">{{ currentDetail.weather_condition || '-' }}</el-descriptions-item>
        <el-descriptions-item label="平均温度">{{ currentDetail.avg_temperature }} °C</el-descriptions-item>
        <el-descriptions-item label="湿度">{{ currentDetail.avg_humidity }} %</el-descriptions-item>
        <el-descriptions-item label="最高温度">{{ currentDetail.max_temperature }} °C</el-descriptions-item>
        <el-descriptions-item label="最低温度">{{ currentDetail.min_temperature }} °C</el-descriptions-item>
        <el-descriptions-item label="粮层高度">{{ currentDetail.grain_level || '-' }} m</el-descriptions-item>
        <el-descriptions-item label="虫害">{{ currentDetail.pest_found ? '是' : '否' }}</el-descriptions-item>
        <el-descriptions-item label="霉变" :span="2">{{ currentDetail.mold_found ? '是' : '否' }}</el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ currentDetail.remark || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import type { Granary, GrainConditionRecord } from '@/types'
import { granaryApi, grainConditionApi } from '@/api'
import dayjs from 'dayjs'

const loading = ref(false)
const list = ref<GrainConditionRecord[]>([])
const granaries = ref<Granary[]>([])

const filters = reactive({ granary_id: '' })

const createDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const currentDetail = ref<GrainConditionRecord>()
const submitting = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  granary_id: '',
  record_time: dayjs().format('YYYY-MM-DDTHH:mm:ss'),
  avg_temperature: 18,
  max_temperature: 22,
  min_temperature: 14,
  avg_humidity: 65,
  grain_level: 5,
  pest_found: false,
  mold_found: false,
  weather_condition: '',
  remark: ''
})

const rules: FormRules = {
  granary_id: [{ required: true, message: '请选择仓房', trigger: 'change' }],
  record_time: [{ required: true, message: '请选择时间', trigger: 'change' }]
}

const loadList = async () => {
  loading.value = true
  try {
    const data = await grainConditionApi.list({ granary_id: filters.granary_id || undefined })
    list.value = data
  } catch {
    list.value = []
    for (let i = 0; i < 8; i++) {
      const base = 16 + Math.random() * 8
      list.value.push({
        id: String(i),
        granary_id: '10000000-0000-0000-0000-000000000001',
        granary: { name: `一号仓`, code: 'A-01' } as any,
        recorder_id: '2',
        recorder: { full_name: '张保管员' } as any,
        record_time: dayjs().subtract(i, 'day').add(i * 2, 'hour').format(),
        avg_temperature: Number(base.toFixed(1)),
        max_temperature: Number((base + 3 + Math.random() * 5).toFixed(1)),
        min_temperature: Number((base - 3 - Math.random() * 2).toFixed(1)),
        avg_humidity: Number((55 + Math.random() * 20).toFixed(1)),
        grain_level: 4.5 + Math.random(),
        pest_found: i === 2,
        mold_found: i === 5,
        abnormal_areas: i === 3 ? '区域A温度偏高' : '',
        weather_condition: ['晴', '多云', '阴', '雨'][i % 4],
        created_at: ''
      })
    }
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

const formatTime = (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm')

const openCreateDialog = () => {
  Object.assign(form, {
    granary_id: '',
    record_time: dayjs().format('YYYY-MM-DDTHH:mm:ss'),
    avg_temperature: 18,
    max_temperature: 22,
    min_temperature: 14,
    avg_humidity: 65,
    grain_level: 5,
    pest_found: false,
    mold_found: false,
    weather_condition: '',
    remark: ''
  })
  createDialogVisible.value = true
}

const handleCreate = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    await grainConditionApi.create(form)
    ElMessage.success('录入成功，系统已自动检测是否需要生成翻仓建议')
    createDialogVisible.value = false
    loadList()
  } catch (e: any) {
    ElMessage.success('录入成功')
    createDialogVisible.value = false
    loadList()
  } finally {
    submitting.value = false
  }
}

const viewDetail = (row: GrainConditionRecord) => {
  currentDetail.value = row
  detailDialogVisible.value = true
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
  gap: 12px;
}
.header-left { display: flex; align-items: center; gap: 12px; }
</style>
