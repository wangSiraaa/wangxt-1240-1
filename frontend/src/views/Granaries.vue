<template>
  <div class="page">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-input
              v-model="filters.keyword"
              placeholder="搜索仓房编号/名称"
              style="width: 240px"
              clearable
              :prefix-icon="Search"
              @keyup.enter="loadList"
              @clear="loadList"
            />
            <el-select v-model="filters.status" placeholder="状态筛选" clearable style="width: 140px" @change="loadList">
              <el-option label="全部" value="" />
              <el-option v-for="(label, key) in GranaryStatusLabels" :key="key" :label="label" :value="key" />
            </el-select>
            <el-button type="primary" :icon="Refresh" @click="loadList">查询</el-button>
          </div>
          <el-button v-if="userStore.hasRole('admin', 'keeper')" type="success" :icon="Plus" @click="openCreateDialog">
            新增仓房
          </el-button>
        </div>
      </template>

      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="code" label="仓房编号" width="100" />
        <el-table-column prop="name" label="仓房名称" width="140" />
        <el-table-column prop="location" label="位置" show-overflow-tooltip />
        <el-table-column label="粮食品种" width="140">
          <template #default="{ row }">
            {{ row.grain_type }} {{ row.grain_variety }}
          </template>
        </el-table-column>
        <el-table-column label="库存/容量" width="160">
          <template #default="{ row }">
            <span>{{ row.grain_weight || 0 }} / {{ row.capacity || 0 }} 吨</span>
            <el-progress
              v-if="row.capacity"
              :percentage="Math.min(Math.round(row.grain_weight / row.capacity * 100), 100)"
              :stroke-width="8"
              style="width: 120px; margin-top: 4px"
            />
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="GranaryStatusColors[row.status] as any" effect="dark">
              {{ GranaryStatusLabels[row.status] }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="保管员" width="120">
          <template #default="{ row }">{{ row.keeper?.full_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="$router.push(`/granaries/${row.id}`)">详情</el-button>
            <el-button type="primary" link size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button
              v-if="userStore.hasRole('admin')"
              type="danger"
              link
              size="small"
              @click="handleDelete(row)"
            >删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑仓房' : '新增仓房'"
      width="640px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="仓房编号" prop="code">
              <el-input v-model="form.code" :disabled="isEdit" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="仓房名称" prop="name">
              <el-input v-model="form.name" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="位置">
              <el-input v-model="form.location" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="容量(吨)">
              <el-input-number v-model="form.capacity" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="当前库存(吨)">
              <el-input-number v-model="form.grain_weight" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="粮食类型">
              <el-select v-model="form.grain_type" placeholder="请选择" style="width: 100%">
                <el-option label="小麦" value="小麦" />
                <el-option label="玉米" value="玉米" />
                <el-option label="稻谷" value="稻谷" />
                <el-option label="大豆" value="大豆" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="品种">
              <el-input v-model="form.grain_variety" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="保管员">
              <el-select v-model="form.keeper_id" placeholder="请选择" filterable style="width: 100%">
                <el-option
                  v-for="k in keepers"
                  :key="k.id"
                  :label="k.full_name"
                  :value="k.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option
                  v-for="(label, key) in GranaryStatusLabels"
                  :key="key"
                  :label="label"
                  :value="key"
                />
              </el-select>
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
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import type { Granary, User, GranaryStatus } from '@/types'
import { GranaryStatusLabels, GranaryStatusColors } from '@/types'
import { granaryApi } from '@/api'
import dayjs from 'dayjs'

const userStore = useUserStore()

const loading = ref(false)
const list = ref<Granary[]>([])
const keepers = ref<User[]>([])

const filters = reactive({
  keyword: '',
  status: '' as GranaryStatus | ''
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  id: '',
  code: '',
  name: '',
  location: '',
  capacity: 0,
  grain_type: '',
  grain_variety: '',
  grain_weight: 0,
  status: 'normal' as GranaryStatus,
  keeper_id: '',
  remark: ''
})

const rules: FormRules = {
  code: [{ required: true, message: '请输入仓房编号', trigger: 'blur' }],
  name: [{ required: true, message: '请输入仓房名称', trigger: 'blur' }]
}

const loadList = async () => {
  loading.value = true
  try {
    const data = await granaryApi.list({
      status: filters.status || undefined,
      keyword: filters.keyword || undefined
    })
    list.value = data
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

const loadKeepers = async () => {
  try {
    keepers.value = await granaryApi.listKeepers()
  } catch {
    keepers.value = []
  }
}

const formatTime = (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm')

const openCreateDialog = () => {
  isEdit.value = false
  Object.assign(form, {
    id: '', code: '', name: '', location: '', capacity: 0, grain_type: '', grain_variety: '', grain_weight: 0, status: 'normal', keeper_id: '', remark: '' })
  dialogVisible.value = true
}

const openEditDialog = (row: Granary) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    code: row.code,
    name: row.name,
    location: row.location || '',
    capacity: row.capacity || 0,
    grain_type: row.grain_type || '',
    grain_variety: row.grain_variety || '',
    grain_weight: row.grain_weight || 0,
    status: row.status,
    keeper_id: row.keeper_id || '',
    remark: row.remark || ''
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    if (isEdit.value) {
      await granaryApi.update(form.id, form)
      ElMessage.success('更新成功')
    } else {
      await granaryApi.create(form)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadList()
  } catch {
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row: Granary) => {
  try {
    await ElMessageBox.confirm(`确定删除仓房 ${row.name} 吗？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await granaryApi.delete(row.id)
    ElMessage.success('删除成功')
    loadList()
  } catch {
  }
}

onMounted(() => {
  loadList()
  loadKeepers()
})
</script>

<style scoped>
.page {
  width: 100%;
}

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
